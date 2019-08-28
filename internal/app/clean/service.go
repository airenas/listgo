package clean

import (
	"net/http"
	"strconv"

	"bitbucket.org/airenas/listgo/internal/app/punctuation/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
)

//Punctuator invokes TF to retrieve punctuation
type Punctuator interface {
	Process(text string) (*api.PResult, error)
}

// ServiceData keeps data required for service work
type ServiceData struct {
	Port   int
	health healthcheck.Handler
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
	ph := cleanHandler{data: data}
	router.Methods("DELETE").Path("/{id}").Handler(&ph)
	if data.health != nil {
		router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
		router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	}
	return router
}

type cleanHandler struct {
	data *ServiceData
}

func (h *cleanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Request from %s", r.RemoteAddr)

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "No ID", http.StatusBadRequest)
		cmdapp.Log.Errorf("No ID")
		return
	}
	cmdapp.Log.Infof("ID: %s", id)
	w.Write([]byte("OK"))
}
