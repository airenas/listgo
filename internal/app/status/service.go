package status

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type serviceMetric struct {
	responseDur  prometheus.ObserverVec
	responseSize prometheus.ObserverVec
}

// ServiceData keeps data required for service work
type ServiceData struct {
	StatusProvider   Provider
	Port             int
	EventChannelFunc eventChannelFunc
	health           healthcheck.Handler

	metrics serviceMetric
}

// StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) error {

	cmdapp.Log.Infof("Listen queue")
	fc := make(chan bool)
	go registerQueue(data, fc, time.Second)

	cmdapp.Log.Infof("Starting HTTP service at %d", data.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Port)
	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	close(fc)
	return nil
}

// NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter()
	sh := promhttp.InstrumentHandlerDuration(data.metrics.responseDur,
		promhttp.InstrumentHandlerResponseSize(data.metrics.responseSize, statusHandler{data: data}))
	router.Methods("GET").Path("/status/{id}").Handler(sh)
	router.Methods("GET").Path("/status").Handler(sh)
	router.Methods("GET").Path("/status/").Handler(sh)
	router.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
	router.Handle("/subscribe", websocketHandler{data: data})
	if data.health != nil {
		router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
		router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	}
	return router
}

type statusHandler struct {
	data *ServiceData
}

type websocketHandler struct {
	data *ServiceData
}

func (h statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Request from %s", r.Host)

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "No ID", http.StatusBadRequest)
		cmdapp.Log.Errorf("No ID")
		return
	}

	result, err := h.data.StatusProvider.Get(id)
	if err != nil {
		http.Error(w, "Cannot get status for ID: "+id, http.StatusBadRequest)
		cmdapp.Log.Errorf("Cannot get status for ID: " + id)
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

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func (h websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("ws request from %s", r.Host)

	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Can not init ws connection", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
	go handleConnection(c)
}
