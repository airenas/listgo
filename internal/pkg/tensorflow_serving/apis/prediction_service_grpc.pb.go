// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.8.0
// source: tensorflow_serving/apis/prediction_service.proto

package apis

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PredictionServiceClient is the client API for PredictionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PredictionServiceClient interface {
	// Classify.
	Classify(ctx context.Context, in *ClassificationRequest, opts ...grpc.CallOption) (*ClassificationResponse, error)
	// Regress.
	Regress(ctx context.Context, in *RegressionRequest, opts ...grpc.CallOption) (*RegressionResponse, error)
	// Predict -- provides access to loaded TensorFlow model.
	Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*PredictResponse, error)
	// MultiInference API for multi-headed models.
	MultiInference(ctx context.Context, in *MultiInferenceRequest, opts ...grpc.CallOption) (*MultiInferenceResponse, error)
	// GetModelMetadata - provides access to metadata for loaded models.
	GetModelMetadata(ctx context.Context, in *GetModelMetadataRequest, opts ...grpc.CallOption) (*GetModelMetadataResponse, error)
}

type predictionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPredictionServiceClient(cc grpc.ClientConnInterface) PredictionServiceClient {
	return &predictionServiceClient{cc}
}

func (c *predictionServiceClient) Classify(ctx context.Context, in *ClassificationRequest, opts ...grpc.CallOption) (*ClassificationResponse, error) {
	out := new(ClassificationResponse)
	err := c.cc.Invoke(ctx, "/tensorflow.serving.PredictionService/Classify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictionServiceClient) Regress(ctx context.Context, in *RegressionRequest, opts ...grpc.CallOption) (*RegressionResponse, error) {
	out := new(RegressionResponse)
	err := c.cc.Invoke(ctx, "/tensorflow.serving.PredictionService/Regress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictionServiceClient) Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*PredictResponse, error) {
	out := new(PredictResponse)
	err := c.cc.Invoke(ctx, "/tensorflow.serving.PredictionService/Predict", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictionServiceClient) MultiInference(ctx context.Context, in *MultiInferenceRequest, opts ...grpc.CallOption) (*MultiInferenceResponse, error) {
	out := new(MultiInferenceResponse)
	err := c.cc.Invoke(ctx, "/tensorflow.serving.PredictionService/MultiInference", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictionServiceClient) GetModelMetadata(ctx context.Context, in *GetModelMetadataRequest, opts ...grpc.CallOption) (*GetModelMetadataResponse, error) {
	out := new(GetModelMetadataResponse)
	err := c.cc.Invoke(ctx, "/tensorflow.serving.PredictionService/GetModelMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PredictionServiceServer is the server API for PredictionService service.
// All implementations must embed UnimplementedPredictionServiceServer
// for forward compatibility
type PredictionServiceServer interface {
	// Classify.
	Classify(context.Context, *ClassificationRequest) (*ClassificationResponse, error)
	// Regress.
	Regress(context.Context, *RegressionRequest) (*RegressionResponse, error)
	// Predict -- provides access to loaded TensorFlow model.
	Predict(context.Context, *PredictRequest) (*PredictResponse, error)
	// MultiInference API for multi-headed models.
	MultiInference(context.Context, *MultiInferenceRequest) (*MultiInferenceResponse, error)
	// GetModelMetadata - provides access to metadata for loaded models.
	GetModelMetadata(context.Context, *GetModelMetadataRequest) (*GetModelMetadataResponse, error)
	mustEmbedUnimplementedPredictionServiceServer()
}

// UnimplementedPredictionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPredictionServiceServer struct {
}

func (UnimplementedPredictionServiceServer) Classify(context.Context, *ClassificationRequest) (*ClassificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Classify not implemented")
}
func (UnimplementedPredictionServiceServer) Regress(context.Context, *RegressionRequest) (*RegressionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Regress not implemented")
}
func (UnimplementedPredictionServiceServer) Predict(context.Context, *PredictRequest) (*PredictResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Predict not implemented")
}
func (UnimplementedPredictionServiceServer) MultiInference(context.Context, *MultiInferenceRequest) (*MultiInferenceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiInference not implemented")
}
func (UnimplementedPredictionServiceServer) GetModelMetadata(context.Context, *GetModelMetadataRequest) (*GetModelMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModelMetadata not implemented")
}
func (UnimplementedPredictionServiceServer) mustEmbedUnimplementedPredictionServiceServer() {}

// UnsafePredictionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PredictionServiceServer will
// result in compilation errors.
type UnsafePredictionServiceServer interface {
	mustEmbedUnimplementedPredictionServiceServer()
}

func RegisterPredictionServiceServer(s grpc.ServiceRegistrar, srv PredictionServiceServer) {
	s.RegisterService(&PredictionService_ServiceDesc, srv)
}

func _PredictionService_Classify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClassificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictionServiceServer).Classify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow.serving.PredictionService/Classify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictionServiceServer).Classify(ctx, req.(*ClassificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PredictionService_Regress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegressionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictionServiceServer).Regress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow.serving.PredictionService/Regress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictionServiceServer).Regress(ctx, req.(*RegressionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PredictionService_Predict_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PredictRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictionServiceServer).Predict(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow.serving.PredictionService/Predict",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictionServiceServer).Predict(ctx, req.(*PredictRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PredictionService_MultiInference_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MultiInferenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictionServiceServer).MultiInference(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow.serving.PredictionService/MultiInference",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictionServiceServer).MultiInference(ctx, req.(*MultiInferenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PredictionService_GetModelMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModelMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictionServiceServer).GetModelMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tensorflow.serving.PredictionService/GetModelMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictionServiceServer).GetModelMetadata(ctx, req.(*GetModelMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PredictionService_ServiceDesc is the grpc.ServiceDesc for PredictionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PredictionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bitbucket.org/airenas/listgo/internal/pkg/tensorflow.serving.PredictionService",
	HandlerType: (*PredictionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Classify",
			Handler:    _PredictionService_Classify_Handler,
		},
		{
			MethodName: "Regress",
			Handler:    _PredictionService_Regress_Handler,
		},
		{
			MethodName: "Predict",
			Handler:    _PredictionService_Predict_Handler,
		},
		{
			MethodName: "MultiInference",
			Handler:    _PredictionService_MultiInference_Handler,
		},
		{
			MethodName: "GetModelMetadata",
			Handler:    _PredictionService_GetModelMetadata_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bitbucket.org/airenas/listgo/internal/pkg/tensorflow_serving/apis/prediction_service.proto",
}
