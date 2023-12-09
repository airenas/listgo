package clean

import (
	"net/http"
	"strconv"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Cleaner deletes information by ID
type Cleaner interface {
	Clean(ID string) error
}

type serviceMetric struct {
	responseDur prometheus.ObserverVec
}

// ServiceData keeps data required for service work
type ServiceData struct {
	Port    int
	health  healthcheck.Handler
	cleaner Cleaner
	metrics serviceMetric
}

// StartWebServer starts the HTTP service and listens for the requests
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

// NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter()
	ch := promhttp.InstrumentHandlerDuration(data.metrics.responseDur, &cleanHandler{data: data})
	router.Methods("DELETE").Path("/{id}").Handler(ch)
	router.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
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
	err := h.data.cleaner.Clean(id)
	if err != nil {
		http.Error(w, "Clean failed", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}
	w.Write([]byte("OK"))
}
