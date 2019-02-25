package result

import (
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	audioFileLoader  FileLoader
	resultFileLoader FileLoader
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
	router.Methods("GET").Path("/result/{id}/{file}").Handler(resultHandler{data: data})
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

	file, err := h.data.audioFileLoader.Load(fileName)
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

type resultHandler struct {
	data *ServiceData
}

func (h resultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("File load request from %s", r.Host)
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "No ID", http.StatusBadRequest)
		cmdapp.Log.Errorf("No ID")
		return
	}
	fileName := mux.Vars(r)["file"]
	if fileName == "" {
		http.Error(w, "No File", http.StatusBadRequest)
		cmdapp.Log.Errorf("No File")
		return
	}

	if strings.Contains(fileName, "..") || strings.Contains(id, "..") {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		cmdapp.Log.Errorf("invalid URL path %s", fileName)
		return
	}

	file, err := h.data.resultFileLoader.Load(id + "/" + fileName)
	if err != nil {
		http.Error(w, "Cannot get file for ID: "+id, http.StatusNotFound)
		cmdapp.Log.Errorf("Cannot get file %s for ID: %s", fileName, id)
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
