package result

import (
	"net/http"
	"strconv"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fileLoader       FileLoader
	fileNameProvider FileNameProvider
	port             int
}

// FileResult - post method response in JSON
type FileResult struct {
	ID string `json:"id"`
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) error {
	cmdapp.Log.Infof("Starting HTTP service at %d", data.port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

//NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/audio/{id}").Handler(audioHandler{data: data})
	return router
}

type audioHandler struct {
	data *ServiceData
}

func (h audioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("File load request from %s", r.Host)
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "No ID", http.StatusBadRequest)
		cmdapp.Log.Errorf("No ID")
		return
	}

	fileName, err := h.data.fileNameProvider.Get(id)
	if err != nil {
		http.Error(w, "Cannot get file for ID: "+id, http.StatusNotFound)
		cmdapp.Log.Errorf("Cannot get file name for ID: " + id)
		return
	}

	file, err := h.data.fileLoader.Load(fileName)
	if err != nil {
		http.Error(w, "Cannot get file for ID: "+id, http.StatusNotFound)
		cmdapp.Log.Errorf("Cannot get file for ID: " + id)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Cannot get file for ID: "+id, http.StatusNotFound)
		cmdapp.Log.Errorf("Cannot get file info for ID: " + id)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileInfo.Name())
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}
