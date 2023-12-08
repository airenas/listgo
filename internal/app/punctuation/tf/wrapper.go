package tf

import (
	"context"
	"strings"

	tf_framework "github.com/airenas/go-tf-serving-protogen/tensorflow/core/framework"
	tf_serving "github.com/airenas/go-tf-serving-protogen/tensorflow_serving/apis"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Wrapper structure used to call TF grpc service
type Wrapper struct {
	url     string
	name    string
	version int
}

// NewWrapper creates Wrapper
func NewWrapper(url string, name string, version int) (*Wrapper, error) {
	res := Wrapper{}
	if strings.TrimSpace(url) == "" {
		return nil, errors.New("No tf.url provided")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("No tf.name provided")
	}
	res.url = url
	res.name = name
	res.version = version
	return &res, nil
}

// Healthy return nil or error is TF model is not accesible
func (w *Wrapper) Healthy() error {
	conn, err := grpc.Dial(w.url, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	r := newModelStatusRequest(w.name, int64(w.version))
	client := tf_serving.NewModelServiceClient(conn)
	st, err := client.GetModelStatus(context.Background(), r)
	if err != nil {
		return err
	}
	for _, s := range st.ModelVersionStatus {
		if s.State == tf_serving.ModelVersionStatus_AVAILABLE {
			return nil
		}
	}
	return errors.New("Model is not available")
}

// Invoke is main method
func (w *Wrapper) Invoke(nums []int32) ([]int32, error) {
	conn, err := grpc.Dial(w.url, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "Cannot connect to the grpc server")
	}
	defer conn.Close()

	r := newPredictRequest(w.name, int64(w.version))
	addInput(r, "word_ids", nums)

	client := tf_serving.NewPredictionServiceClient(conn)
	resp, err := client.Predict(context.Background(), r)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot invoke tf server")
	}

	out := resp.GetOutputs()
	for _, v := range out {
		d := v.GetTensorShape().GetDim()
		if len(d) < 3 {
			return nil, errors.Wrapf(err, "Expected result dimmention 3, got %v", d)
		}
		return makeRes(v.GetFloatVal(), int(d[1].Size), int(d[2].Size)), nil
	}

	return nil, errors.New("No result")
}

func makeRes(in []float32, d1, d2 int) []int32 {
	result := make([]int32, d1)
	for i := 0; i < d1; i++ {
		result[i] = argmax(in[i*d2 : (i+1)*d2])
	}
	return result
}

func argmax(in []float32) int32 {
	r := 0
	m := in[0]
	for i := 1; i < len(in); i++ {
		if m < in[i] {
			m = in[i]
			r = i
		}
	}
	return int32(r)
}

func newPredictRequest(modelName string, modelVersion int64) *tf_serving.PredictRequest {
	return &tf_serving.PredictRequest{
		ModelSpec: &tf_serving.ModelSpec{
			Name: modelName,
		},
		Inputs: make(map[string]*tf_framework.TensorProto),
	}
}

func newModelStatusRequest(modelName string, modelVersion int64) *tf_serving.GetModelStatusRequest {
	return &tf_serving.GetModelStatusRequest{
		ModelSpec: &tf_serving.ModelSpec{
			Name: modelName,
		}}
}

func addInput(pr *tf_serving.PredictRequest, tensorName string, data []int32) (err error) {
	tp := &tf_framework.TensorProto{
		Dtype: tf_framework.DataType_DT_INT32,
		TensorShape: &tf_framework.TensorShapeProto{
			Dim: []*tf_framework.TensorShapeProto_Dim{
				{Size: 1},
				{Size: int64(len(data))},
			},
		},
		IntVal: data,
	}
	pr.Inputs[tensorName] = tp
	return
}
