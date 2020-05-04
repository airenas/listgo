package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestWrongPath(t *testing.T) {
	testCode(t, newTestData(), httptest.NewRequest("GET", "/invalid", nil), 404)
	testCode(t, newTestData(), httptest.NewRequest("GET", "/olia", nil), 404)
}

func TestReturnsGet(t *testing.T) {
	testCode(t, newTestData(), httptest.NewRequest("GET", "/metrics", nil), 200)
	testCode(t, newTestData(), httptest.NewRequest("GET", "/live", nil), 200)
	testCode(t, newTestData(), httptest.NewRequest("GET", "/ready", nil), 200)
}

func TestMetricsPost(t *testing.T) {
	r := request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: time.Now().UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
}

func TestMetricsPostFails(t *testing.T) {
	r := request{ID: "", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: time.Now().UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
	r = request{ID: "id", Model: "m", Task: "", Worker: "w", Type: "start", Timestap: time.Now().UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
	r = request{ID: "id", Model: "m", Task: "t", Worker: "", Type: "start", Timestap: time.Now().UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "st", Timestap: time.Now().UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: time.Now().Add(time.Hour).UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: time.Now().Add(-5 * time.Hour).UnixNano()}
	testCode(t, newTestData(), httptest.NewRequest("POST", "/metrics", encode(&r)), 400)
}

func TestMetricsAddsStart(t *testing.T) {
	n := time.Now().UnixNano()
	r := request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: n}
	d := newTestData()
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	assert.Equal(t, n, d.dMap["w:t:m"]["id"].timestap)
	assert.Equal(t, 0, testutil.CollectAndCount(d.tasksMetrics))
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksStarted))
	assert.Equal(t, 0, testutil.CollectAndCount(d.tasksEnded))
}

func TestMetricsAddsEnd(t *testing.T) {
	n := time.Now().UnixNano()
	r := request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: n}
	d := newTestData()
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	ne := n + time.Minute.Nanoseconds()
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "end", Timestap: ne}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	assert.Nil(t, d.dMap["w:t:m"]["id"])
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksMetrics))
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksStarted))
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksEnded))
}

func TestMetricsAddsSecond(t *testing.T) {
	n := time.Now().UnixNano()
	r := request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: n}
	d := newTestData()
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	ne := n + time.Minute.Nanoseconds()
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "end", Timestap: ne}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	r = request{ID: "id", Model: "m2", Task: "t", Worker: "w", Type: "start", Timestap: n}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	r = request{ID: "id", Model: "m2", Task: "t", Worker: "w", Type: "end", Timestap: ne}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	assert.Equal(t, 2, testutil.CollectAndCount(d.tasksMetrics))
	assert.Equal(t, 2, testutil.CollectAndCount(d.tasksStarted))
	assert.Equal(t, 2, testutil.CollectAndCount(d.tasksEnded))
}

func TestMetricsGroupById(t *testing.T) {
	n := time.Now().UnixNano()
	r := request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: n}
	d := newTestData()
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	ne := n + time.Minute.Nanoseconds()
	r = request{ID: "id", Model: "m", Task: "t", Worker: "w", Type: "end", Timestap: ne}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	r = request{ID: "id2", Model: "m", Task: "t", Worker: "w", Type: "start", Timestap: n}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	r = request{ID: "id2", Model: "m", Task: "t", Worker: "w", Type: "end", Timestap: ne}
	testCode(t, d, httptest.NewRequest("POST", "/metrics", encode(&r)), 200)
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksMetrics))
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksStarted))
	assert.Equal(t, 1, testutil.CollectAndCount(d.tasksEnded))
}

func testCode(t *testing.T, data *ServiceData, req *http.Request, code int) {
	resp := httptest.NewRecorder()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
}

func newTestData() *ServiceData {
	data, _ := newServiceData()
	data.health = healthcheck.NewHandler()
	return data
}

func encode(d *request) *bytes.Buffer {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.Encode(d)
	return b
}
