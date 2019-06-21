package upload

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/badoux/checkmail"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/heptiolabs/healthcheck"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	FileSaver     FileSaver
	MessageSender messages.Sender
	StatusSaver   status.Saver
	RequestSaver  RequestSaver
	Port          int
	health        healthcheck.Handler
}

// FileResult - post method response in JSON
type FileResult struct {
	ID string `json:"id"`
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) error {
	cmdapp.Log.Infof("Starting HTTP service at %d", data.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

//NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("POST").Path("/upload").Handler(uploadHandler{data: data})
	router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
	router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	return router
}

type uploadHandler struct {
	data *ServiceData
}

func (h uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Saving file from %s", r.Host)

	r.ParseMultipartForm(32 << 20)
	externalID := r.FormValue("externalID")

	email := r.FormValue("email")
	if email != "" {
		err := checkmail.ValidateFormat(email)
		if err != nil {
			http.Error(w, "Wrong email", http.StatusBadRequest)
			cmdapp.Log.Errorf("Wrong email")
			return
		}
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(handler.Filename)
	ext = strings.ToLower(ext)
	if !checkFileExtension(ext) {
		http.Error(w, "Wrong file extension: "+ext, http.StatusBadRequest)
		cmdapp.Log.Errorf("Wrong file extension: " + ext)
		return
	}

	id := uuid.New().String()
	fileName := id + ext

	err = h.data.RequestSaver.Save(api.RequestData{ID: id, Email: email, File: fileName, ExternalID: externalID})
	if err != nil {
		http.Error(w, "Can not save request to DB", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	err = h.data.StatusSaver.Save(id, status.Uploaded)
	if err != nil {
		http.Error(w, "Can not save status", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	err = h.data.FileSaver.Save(fileName, file)
	if err != nil {
		http.Error(w, "Can not save file", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	err = h.data.MessageSender.Send(messages.NewQueueMessage(id), messages.Decode, "")
	if err != nil {
		http.Error(w, "Can not send decode message", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	result := FileResult{id}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&result)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}
}

func checkFileExtension(ext string) bool {
	return ext == ".wav" || ext == ".mp3" || ext == ".mp4"
}
