package status

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	StatusProvider   Provider
	Port             int
	EventChannelFunc eventChannelFunc
}

//StartWebServer starts the HTTP service and listens for the requests
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

//NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter()
	router.Methods("GET").Path("/result/{id}").Handler(statusHandler{data: data})
	router.Methods("GET").Path("/result").Handler(statusHandler{data: data})
	router.Methods("GET").Path("/result/").Handler(statusHandler{data: data})
	router.Handle("/subscribe", websocketHandler{data: data})
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
		cmdapp.Log.Error(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultBytes)
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func (h websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("ws request from %s", r.Host)

	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		setError(w, "Can not init ws connection", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
	go handleConnection(c)
}

func setError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
