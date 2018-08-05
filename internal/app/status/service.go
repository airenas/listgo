package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	StatusProvider Provider
	Port           int
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
	router.Methods("GET").Path("/result/{id}").Handler(statusHandler{data: data})
	return router
}

type statusHandler struct {
	data *ServiceData
}

func (h statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Request from %s", r.Host)

	id := mux.Vars(r)["id"]
	if id == "" {
		setError(w, "No ID", http.StatusBadRequest)
		cmdapp.Log.Errorf("No ID")
		return
	}

	result, err := h.data.StatusProvider.Get(id)
	if err != nil {
		setError(w, "Cannot get status for ID: "+id, http.StatusBadRequest)
		cmdapp.Log.Errorf("Cannot get status for ID: " + id)
		return
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		setError(w, "Can not prepare result", http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultBytes)
}

func setError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
