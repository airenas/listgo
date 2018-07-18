package upload

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	FileSaver     FileSaver
	MessageSender MessageSender
	Port          string
}

// FileResult - post method response in JSON
type FileResult struct {
	ID string `json:"id"`
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) {
	log.Println("Starting HTTP service at " + data.Port)
	r := NewRouter(data)
	http.Handle("/", r)

	err := http.ListenAndServe(":"+data.Port, nil)

	if err != nil {
		log.Println("An error occured starting HTTP listener at port " + data.Port)
		log.Println("Error: " + err.Error())
	}
}

//NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("POST").Path("/upload").Handler(uploadHandler{data: data})
	return router
}

type uploadHandler struct {
	data *ServiceData
}

func (h uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Saving file from " + r.Host)

	r.ParseMultipartForm(32 << 20)
	email := r.FormValue("email")
	if email == "" {
		setError(w, "No email", http.StatusBadRequest)
		log.Println("No email")
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		setError(w, "No file", http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer file.Close()
	id := uuid.New().String()

	ext := filepath.Ext(handler.Filename)
	fileName := id + ext

	err = h.data.FileSaver.Save(fileName, file)
	if err != nil {
		setError(w, "Can not save file", http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = h.data.MessageSender.Send(createDecodeMsg(id))
	if err != nil {
		setError(w, "Can not send decode message", http.StatusBadRequest)
		log.Println(err)
		return
	}

	result := FileResult{id}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		setError(w, "Can not prepare result", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultBytes)
}

func createDecodeMsg(id string) msgsender.Message {
	return msgsender.Message{ID: id, Queue: "Decode"}
}

func setError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
