package upload

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/listgo/internal/app/upload/api"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/persistence"
	"github.com/airenas/listgo/internal/pkg/status"
	"github.com/airenas/listgo/internal/pkg/utils"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/badoux/checkmail"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/heptiolabs/healthcheck"
)

type serviceMetric struct {
	uploadResponseDur prometheus.ObserverVec
	uploadRequestSize prometheus.ObserverVec

	recResponseDur prometheus.ObserverVec
}

// ServiceData keeps data required for service work
type ServiceData struct {
	FileSaver          FileSaver
	MessageSender      messages.Sender
	StatusSaver        status.Saver
	RequestSaver       RequestSaver
	RecognizerMap      RecognizerMap
	RecognizerProvider RecognizerProvider

	Port    int
	health  healthcheck.Handler
	metrics serviceMetric
}

// FileResult - post method response in JSON
type FileResult struct {
	ID string `json:"id"`
}

// StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *ServiceData) error {
	cmdapp.Log.Infof("Starting HTTP service at %d", data.Port)
	r := NewRouter(data)

	portStr := strconv.Itoa(data.Port)
	srv := http.Server{
		Addr:              ":" + portStr,
		WriteTimeout:      300 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       300 * time.Second,
		IdleTimeout:       10 * time.Minute,
		Handler:           r,
	}

	w := cmdapp.Log.Writer()
	defer w.Close()
	l := log.New(w, "", 0)
	gracehttp.SetLogger(l)

	return gracehttp.Serve(&srv)
}

// NewRouter creates the router for HTTP service
func NewRouter(data *ServiceData) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	uh := promhttp.InstrumentHandlerDuration(data.metrics.uploadResponseDur,
		promhttp.InstrumentHandlerRequestSize(data.metrics.uploadRequestSize, uploadHandler{data: data}))
	rh := promhttp.InstrumentHandlerDuration(data.metrics.recResponseDur, recognizersHandler{data: data})
	router.Methods("POST").Path("/upload").Handler(uh)
	router.Methods("GET").Path("/recognizers").Handler(rh)
	router.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
	router.Methods("GET").Path("/live").HandlerFunc(data.health.LiveEndpoint)
	router.Methods("GET").Path("/ready").HandlerFunc(data.health.ReadyEndpoint)
	return router
}

type uploadHandler struct {
	data *ServiceData
}

func (h uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Saving file from %s", r.Host)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Can't parse MultipartForm", http.StatusBadRequest)
		cmdapp.Log.Error(errors.Wrap(err, "Can't parse MultipartForm"))
		return
	}
	defer cleanFiles(r.MultipartForm)
	err = validateFormParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
	externalID := r.FormValue(api.PrmExternalID)
	numberOfSpeakers := r.FormValue(api.PrmNumberOfSpeakers)
	skipNumJoin := r.FormValue(api.PrmSkipNumJoin)
	sepSpOnCh := r.FormValue(api.PrmSepSpeakersOnChannel)
	email := r.FormValue(api.PrmEmail)
	if email != "" {
		err := checkmail.ValidateFormat(email)
		if err != nil {
			http.Error(w, "Wrong email", http.StatusBadRequest)
			cmdapp.Log.Errorf("Wrong email")
			return
		}
	}

	recognizer := r.FormValue(api.PrmRecognizer)
	recID, err := h.data.RecognizerMap.Get(recognizer)
	if err != nil {
		if err == api.ErrRecognizerNotFound {
			http.Error(w, getRecErrMsg(recognizer), http.StatusBadRequest)
		} else {
			http.Error(w, "Can't select recognizer", http.StatusInternalServerError)
		}
		cmdapp.Log.Errorf("Problem with recognizer '%s'. %s", recognizer, err.Error())
		return
	}
	cmdapp.Log.Infof("Found recognizer '%s' for '%s'", recID, recognizer)

	files, fHeaders, err := takeFiles(r, api.PrmFile)
	for _, f := range files {
		fInt := f
		defer fInt.Close()
	}
	if err != nil && len(files) == 0 {
		http.Error(w, "No file", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}
	if err != nil {
		http.Error(w, "Wrong input form", http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}

	err = validateFiles(fHeaders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		cmdapp.Log.Error(err)
		return
	}

	if len(files) > 1 && utils.ParamTrue(sepSpOnCh) {
		http.Error(w, "'sepSpeakersOnChannel' not supported with multiple files", http.StatusBadRequest)
		cmdapp.Log.Error("'sepSpeakersOnChannel' not supported with multiple files")
		return
	}

	id := uuid.New().String()
	fileName := id + ".mp3"
	audioReady := false
	if len(files) == 1 {
		ext := filepath.Ext(fHeaders[0].Filename)
		ext = strings.ToLower(ext)
		fileName = id + ext
		audioReady = true
	}

	err = h.data.RequestSaver.Save(&persistence.Request{ID: id, Email: email, File: fileName, ExternalID: externalID,
		RecognizerKey: recognizer, RecognizerID: recID})
	if err != nil {
		http.Error(w, "Can not save request to DB", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	err = h.data.StatusSaver.SaveF(id, map[string]interface{}{
		"status":                 status.Name(status.Uploaded),
		persistence.StAudioReady: audioReady}, nil)
	if err != nil {
		http.Error(w, "Can not save status", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	err = saveFiles(h.data.FileSaver, id, files, fHeaders)
	if err != nil {
		http.Error(w, "Can not save file", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	tags := []messages.Tag{messages.NewTag(messages.TagNumberOfSpeakers, numberOfSpeakers),
		messages.NewTag(messages.TagTimestamp, strconv.FormatInt(time.Now().Unix(), 10))}
	if skipNumJoin != "" {
		tags = append(tags, messages.NewTag(messages.TagSkipNumJoin, skipNumJoin))
	}
	if utils.ParamTrue(sepSpOnCh) {
		tags = append(tags, messages.NewTag(messages.TagSepSpeakersOnChannel, "1"))
	}

	msg := messages.Decode
	if len(files) > 1 {
		msg = messages.DecodeMultiple
	}

	err = h.data.MessageSender.Send(messages.NewQueueMessage(id, recID, tags), msg, "")
	if err != nil {
		http.Error(w, "Can not send decode message", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	result := FileResult{id}
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&result)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}
}

func cleanFiles(f *multipart.Form) {
	if f != nil {
		f.RemoveAll()
	}
}

func getRecErrMsg(rec string) string {
	if rec == "" {
		return "No recognizer"
	}
	return "Unknown recognizer: " + rec
}

type recognizersHandler struct {
	data *ServiceData
}

func (h recognizersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cmdapp.Log.Infof("Recognizers get %s", r.Host)
	rec, err := h.data.RecognizerProvider.GetAll()
	if err != nil {
		http.Error(w, "Can not get recognizers", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(&rec)
	if err != nil {
		http.Error(w, "Can not prepare result", http.StatusInternalServerError)
		cmdapp.Log.Error(err)
		return
	}
}

func validateFormParams(r *http.Request) error {
	form := r.Form
	allowed := map[string]bool{api.PrmEmail: true, api.PrmRecognizer: true, api.PrmExternalID: true,
		api.PrmNumberOfSpeakers: true, api.PrmSkipNumJoin: true, api.PrmSepSpeakersOnChannel: true}
	for k := range form {
		_, f := allowed[k]
		if !f {
			return errors.Errorf("Unknown parameter '%s'", k)
		}
	}
	for _, p := range [...]string{api.PrmNumberOfSpeakers, api.PrmSkipNumJoin} {
		if err := validateInjection(r, p); err != nil {
			return err
		}
	}
	return validateFormFiles(r.MultipartForm)
}

func validateFormFiles(form *multipart.Form) error {
	check := make(map[string]bool)
	if form != nil {
		for k := range form.File {
			check[k] = true
		}
	}
	if !check[api.PrmFile] {
		return errors.New("no form file parameter 'file'")
	}
	delete(check, api.PrmFile)
	for i := 2; i <= 10; i++ {
		pn := api.PrmFile + strconv.Itoa(i)
		if !check[pn] {
			break
		}
		delete(check, pn)
	}
	for k := range check {
		return errors.Errorf("unexpected form file parameters '%v'", k)
	}
	return nil
}

func validateInjection(r *http.Request, paramName string) error {
	p := r.FormValue(paramName)
	lp := strings.ToLower(p)
	for _, k := range []string{"$", "(", ")", "eval", "shell", "{", "}"} {
		if strings.Contains(lp, k) {
			return errors.Errorf("Wrong parameter '%s' value '%s'", paramName, p)
		}
	}
	return nil
}

func takeFiles(r *http.Request, paramName string) ([]multipart.File, []*multipart.FileHeader, error) {
	file, handler, err := r.FormFile(paramName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "no form param file")
	}
	fRes := []multipart.File{file}
	fhRes := []*multipart.FileHeader{handler}
	for i := 2; i <= 10; i++ {
		file, handler, err := r.FormFile(paramName + strconv.Itoa(i))
		if err == http.ErrMissingFile {
			break
		}
		if err != nil {
			return fRes, nil, errors.Wrapf(err, "error reading form param %s", paramName+strconv.Itoa(i))
		}
		fRes = append(fRes, file)
		fhRes = append(fhRes, handler)
	}
	return fRes, fhRes, nil
}

func validateFiles(fHeaders []*multipart.FileHeader) error {
	for _, h := range fHeaders {
		ext := filepath.Ext(h.Filename)
		if !utils.SupportAudioExt(strings.ToLower(ext)) {
			return errors.New("wrong file extension: " + ext)
		}
		if strings.Contains(h.Filename, "..") {
			return errors.New("wrong file name: " + h.Filename)
		}
	}
	return nil
}

func saveFiles(fs FileSaver, id string, files []multipart.File, fHeaders []*multipart.FileHeader) error {
	if len(files) == 1 {
		ext := filepath.Ext(fHeaders[0].Filename)
		ext = strings.ToLower(ext)
		return fs.Save(id+ext, files[0])
	}

	for i, f := range files {
		fn := fHeaders[i].Filename
		if fn == "" {
			return errors.New("no file name in multipart")
		}
		fn = filepath.Join(id, sanitizeName(fn))
		err := fs.Save(toLowerExt(fn), f)
		if err != nil {
			return errors.Wrapf(err, "can't save %s", fn)
		}
	}
	return nil
}

func sanitizeName(s string) string {
	res := strings.ReplaceAll(s, " ", "_")
	return filepath.Base(res)
}

func toLowerExt(f string) string {
	ext := filepath.Ext(f)
	return f[:len(f)-len(ext)] + strings.ToLower(ext)
}
