package tf

import (
	"context"
	"strings"

	framework "bitbucket.org/airenas/listgo/internal/pkg/tensorflow/core/framework"
	pb "bitbucket.org/airenas/listgo/internal/pkg/tensorflow_serving/apis"
	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)
		
// Wrapper structure used to call TF grpc service
type Wrapper struct {
	url     string
	name    string
	version int
}

//NewWrapper creates Wrapper
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

//Invoke is main method
func (w *Wrapper) Invoke(nums []int32) ([]int32, error) {

	conn, err := grpc.Dial(w.url, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "Cannot connect to the grpc server")
	}
	defer conn.Close()

	r := newPredictRequest(w.name, int64(w.version))
	addInput(r, "word_ids", nums, []int64{1, int64(len(nums))})

	client := pb.NewPredictionServiceClient(conn)
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

func newPredictRequest(modelName string, modelVersion int64) (pr *pb.PredictRequest) {
	return &pb.PredictRequest{
		ModelSpec: &pb.ModelSpec{
			Name: modelName,
			Version: &google_protobuf.Int64Value{
				Value: modelVersion,
			},
		},
		Inputs: make(map[string]*framework.TensorProto),
	}
}

// if tensor is one dim, shapeSize is nil
func addInput(pr *pb.PredictRequest, tensorName string, data []int32, shapeSize []int64) (err error) {
	tp := &framework.TensorProto{
		Dtype: framework.DataType_DT_INT32,
	}
	tp.IntVal = data
	tp.TensorShape = &framework.TensorShapeProto{
		Dim: []*framework.TensorShapeProto_Dim{},
	}
	for _, size := range shapeSize {
		tp.TensorShape.Dim = append(tp.TensorShape.Dim,
			&framework.TensorShapeProto_Dim{
				Size: size,
				Name: "",
			})
	}
	pr.Inputs[tensorName] = tp
	return
}
