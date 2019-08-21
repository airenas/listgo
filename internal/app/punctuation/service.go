package punctuation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
)

//Punctuator invokes TF to retrieve punctuation
type Punctuator interface {
	Process(text string) (string, error)
}

// ServiceData keeps data required for service work
type ServiceData struct {
	Port       int
	health     healthcheck.Handler
	punctuator Punctuator
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
	router := mux.NewRouter()
	ph := punctuationHandler{data: data}
	router.Methods("POST").Path("/punctuation").Handler(&ph)
	router.Methods("POST").Path("/punctuation/").Handler(&ph)
	if data.health != nil {
		router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
		router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	}
	return router
}

type punctuationHandler struct {
	data *ServiceData
}

func (h *punctuationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Request from %s", r.Host)

	decoder := json.NewDecoder(r.Body)
	var input Input
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		cmdapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	if strings.TrimSpace(input.Text) == "" {
		http.Error(w, "No text", http.StatusBadRequest)
		cmdapp.Log.Error("No text")
		return
	}

	result := Output{}
	result.Original = input.Text
	result.Punctuated, err = h.data.punctuator.Process(input.Text)
	if err != nil {
		http.Error(w, "Cannot punctuate", http.StatusInternalServerError)
		cmdapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&result)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
}
