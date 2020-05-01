package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	Port   int
	health healthcheck.Handler

	metricDur    *prometheus.HistogramVec
	tasksMetrics *prometheus.HistogramVec
	dMap         map[string]map[string]*startTime
	lock         *sync.Mutex
}

func newServiceData() (*ServiceData, error) {
	res := &ServiceData{}
	res.lock = &sync.Mutex{}
	res.dMap = make(map[string]map[string]*startTime)
	err := initMetrics(res)
	if err != nil {
		return nil, errors.Wrap(err, "Can' int metrics")
	}
	return res, nil
}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) error {
	cmdapp.Log.Infof("Starting HTTP service at %d", data.Port)
	r := NewRouter(data)
	http.Handle("/", r)
	portStr := strconv.Itoa(data.Port)

	go checkForExpired(data)

	err := http.ListenAndServe(":"+portStr, nil)

	if err != nil {
		return errors.Wrap(err, "Can't start HTTP listener at port "+portStr)
	}
	return nil
}

//NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter()
	mh := promhttp.InstrumentHandlerDuration(data.metricDur, &metricsHandler{data: data})
	router.Methods("POST").Path("/metrics").Handler(mh)
	router.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
	if data.health != nil {
		router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
		router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	}
	return router
}

type request struct {
	ID       string `json:"id"`
	Timestap int64  `json:"timestamp"`
	Type     string `json:"type"`
	Worker   string `json:"worker"`
	Task     string `json:"task"`
}

type startTime struct {
	timestap int64
	added    time.Time
}

type metricsHandler struct {
	data *ServiceData
}

func (h *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Debugf("Request from %s", r.Host)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var inData request
	err := dec.Decode(&inData)
	if err != nil {
		cmdapp.Log.Errorf("Bad input. ", err)
		http.Error(w, "Bad input", http.StatusBadRequest)
		return
	}
	err = validate(&inData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}

	h.data.lock.Lock()
	defer h.data.lock.Unlock()

	key := inData.Worker + "_t:" + inData.Task
	if inData.Type == "start" {
		idMap, f := h.data.dMap[inData.Worker+"_t:"+inData.Task]
		if !f {
			idMap = make(map[string]*startTime)
			h.data.dMap[inData.Worker+"_t:"+inData.Task] = idMap
		}
		idMap[inData.ID] = &startTime{timestap: inData.Timestap, added: time.Now()}
	}
	if inData.Type == "end" {
		idMap, f := h.data.dMap[key]
		if !f {
			cmdapp.Log.Warn("No started task found for " + key)
			return
		}
		startData, f := idMap[inData.ID]
		if !f {
			cmdapp.Log.Warnf("No started task found for %s, ID: %s", key, inData.ID)
			return
		}
		if startData.timestap > inData.Timestap {
			http.Error(w, "Wrong end timestamp", http.StatusBadRequest)
			return
		}
		addMetric(h.data, &inData, startData)
		delete(idMap, inData.ID)
	}
}

func validate(inData *request) error {
	if inData.Worker == "" {
		return errors.New("No worker")
	}
	if !(inData.Type == "start" || inData.Type == "end") {
		return errors.New("Type must be start or end")
	}
	if inData.Task == "" {
		return errors.New("No task")
	}
	if inData.ID == "" {
		return errors.New("No ID")
	}
	if inData.Timestap < time.Now().Add(-2*time.Hour).UnixNano() ||
		inData.Timestap > time.Now().Add(time.Minute).UnixNano() {
		return errors.New("Wrong timestamp")
	}
	return nil
}

func addMetric(data *ServiceData, en *request, st *startTime) {
	data.tasksMetrics.
		With(prometheus.Labels{"worker": en.Worker, "task": en.Task}).
		Observe(float64(en.Timestap-st.timestap) / float64(1e9))
	time.Now().UnixNano()
}

func checkForExpired(data *ServiceData) {
	for {
		time.Sleep(time.Hour)
		checkForExpiredInt(data, time.Now().Add(-3*time.Hour))
	}
}

func checkForExpiredInt(data *ServiceData, exp time.Time) {
	data.lock.Lock()
	defer data.lock.Unlock()
	for kTask, vTask := range data.dMap {
		for k, v := range vTask {
			if v.added.Before(exp) {
				delete(vTask, k)
			}
		}
		if len(vTask) == 0 {
			cmdapp.Log.Debugf("Delete expired key %s", kTask)
			delete(data.dMap, kTask)
		}
	}
}
