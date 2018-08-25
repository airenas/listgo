package status

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var idConnectionMap = make(map[string]*websocket.Conn)
var connectionIDMap = make(map[*websocket.Conn]string)
var mapLock = sync.Mutex{}

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
	go registerQueue(data)

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

func handleConnection(conn *websocket.Conn) {
	defer deleteConnection(conn)
	for {
		cmdapp.Log.Infof("handleConnection")
		_, message, err := conn.ReadMessage()
		if err != nil {
			cmdapp.Log.Error(err)
			break
		} else {
			saveConnection(conn, string(message))
		}
	}
	cmdapp.Log.Infof("handleConnection finish")
}

func deleteConnection(conn *websocket.Conn) {
	cmdapp.Log.Infof("deleteConnection")
	mapLock.Lock()
	defer mapLock.Unlock()
	defer conn.Close()
	cmdapp.Log.Info("delete connection")
	id, found := connectionIDMap[conn]
	if found {
		delete(idConnectionMap, id)
	}
	delete(connectionIDMap, conn)
	cmdapp.Log.Infof("deleteConnection finish: %d", len(connectionIDMap))
}

func saveConnection(conn *websocket.Conn, id string) {
	cmdapp.Log.Infof("saveConnection")
	mapLock.Lock()
	defer mapLock.Unlock()
	cmdapp.Log.Info("delete connection" + id)
	idOld, found := connectionIDMap[conn]
	if found {
		delete(idConnectionMap, idOld)
	}
	connectionIDMap[conn] = id
	idConnectionMap[id] = conn
	cmdapp.Log.Infof("saveConnection finish: %d", len(connectionIDMap))
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

func registerQueue(data *ServiceData) {
	fc := make(chan bool)
	wait := time.Duration(1)
	for {
		cmdapp.Log.Infof("Trying listening queue")
		msgs, err := data.EventChannelFunc()
		if err != nil {
			cmdapp.Log.Error(err)
			wait = wait * 2
			if wait > 60 {
				wait = 60
			}
			cmdapp.Log.Infof("Wait before reconnect %d s", wait)
			time.Sleep(wait * time.Second)
			continue
		}
		wait = 1
		go listenQueue(msgs, data, fc)
		<-fc
	}
}

func processMsg(d *amqp.Delivery, data *ServiceData) error {
	id := string(d.Body)
	cmdapp.Log.Infof("processMsg event " + id)
	conn, found := idConnectionMap[id]
	if found {
		result, err := data.StatusProvider.Get(id)
		if err != nil {
			return errors.Wrap(err, "Cannot get status for ID: "+id)
		}
		err = conn.WriteJSON(result)
		if err != nil {
			return errors.Wrap(err, "Cannot write to websockket: "+id)
		}
	} else {
		cmdapp.Log.Infof("not found " + id)
	}
	return nil
}
