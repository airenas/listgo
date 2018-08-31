package status

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type eventChannelFunc func() (<-chan amqp.Delivery, error)

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
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET").Path("/result/{id}").Handler(statusHandler{data: data})
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

func listenQueue(channel <-chan amqp.Delivery, data *ServiceData, fc chan<- bool) {
	for d := range channel {
		err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Errorf("Can't process message %s\n%s", d.MessageId, string(d.Body))
			cmdapp.Log.Error(err)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	fc <- true
}

func registerQueue(data *ServiceData, quitChan <-chan bool, initialWait time.Duration) {
	fc := make(chan bool)
	wait := initialWait
	for {
		select {
		case <-quitChan:
			cmdapp.Log.Infof("Quit listening queue")
			return
		default:
			cmdapp.Log.Infof("Trying listening queue")
			msgs, err := data.EventChannelFunc()
			if err != nil {
				cmdapp.Log.Error(err)
				wait = wait * 2
				if wait > time.Minute {
					wait = time.Minute
				}
				cmdapp.Log.Infof("Wait before reconnect %d s", wait/time.Second)
				time.Sleep(wait)
				continue
			}
			wait = initialWait
			go listenQueue(msgs, data, fc)
			<-fc
		}
	}
}

func processMsg(d *amqp.Delivery, data *ServiceData) error {
	id := string(d.Body)
	cmdapp.Log.Infof("processMsg event " + id)
	conns, found := getConnections(id)
	if found {
		result, err := data.StatusProvider.Get(id)
		if err != nil {
			return errors.Wrap(err, "Cannot get status for ID: "+id)
		}
		for c := range conns {
			sendMsg(c, result)
		}
	} else {
		cmdapp.Log.Infof("not found " + id)
	}
	return nil
}

func sendMsg(c wsConn, result *api.TranscriptionResult) error {
	cmdapp.Log.Infof("sending result for %s", result.ID)
	conn, ok := c.(*websocket.Conn)
	if ok {
		err := conn.WriteJSON(result)
		cmdapp.Log.Infof("sent")
		if err != nil {
			cmdapp.Log.Error("Cannot write to websockket")
			cmdapp.Log.Error(err)
		}
	} else {
		cmdapp.Log.Errorf("Can not cast to *websocket.Conn")
	}
	return nil
}
