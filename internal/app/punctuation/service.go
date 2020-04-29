package punctuation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"bitbucket.org/airenas/listgo/internal/app/punctuation/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
)

//Punctuator invokes TF to retrieve punctuation
type Punctuator interface {
	Process(data []string) (*api.PResult, error)
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
	pAh := punctuationArrayHandler{data: data}
	router.Methods("POST").Path("/punctuation").Handler(&ph)
	router.Methods("POST").Path("/punctuation/").Handler(&ph)
	router.Methods("POST").Path("/punctuationArray").Handler(&pAh)
	router.Methods("POST").Path("/punctuationArray/").Handler(&pAh)
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

	queryValues := r.URL.Query()
	debug := queryValues.Get("debug")

	decoder := json.NewDecoder(r.Body)
	var input Input
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		cmdapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	process(h.data, convertToArray(input.Text), debug == "1", w)
}

type punctuationArrayHandler struct {
	data *ServiceData
}

func (h *punctuationArrayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Request from %s", r.Host)

	queryValues := r.URL.Query()
	debug := queryValues.Get("debug")

	decoder := json.NewDecoder(r.Body)
	var input InputArray
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, "Cannot decode input", http.StatusBadRequest)
		cmdapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	process(h.data, input.Words, debug == "1", w)
}

func process(data *ServiceData, words []string, debug bool, w http.ResponseWriter) {
	if len(words) == 0 {
		http.Error(w, "No input text", http.StatusBadRequest)
		cmdapp.Log.Error("No input text")
		return
	}

	result := Output{}
	result.Original = words
	pr, err := data.punctuator.Process(words)
	if err != nil {
		http.Error(w, "Cannot punctuate", http.StatusInternalServerError)
		cmdapp.Log.Error("Cannot decode input" + err.Error())
		return
	}

	result.Punctuated = pr.Punctuated
	if debug {
		result.WordIDs = pr.WordIDs
		result.PunctIDs = pr.PunctIDs
	}
	result.PunctuatedText = pr.PunctuatedText

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&result)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
	}
}

func convertToArray(strs string) []string {
	arr := strings.Split(strings.TrimSpace(strs), " ")
	result := make([]string, 0)
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
