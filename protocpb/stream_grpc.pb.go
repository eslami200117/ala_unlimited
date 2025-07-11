// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: protocpb/stream.proto

package protocpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	PriceService_StreamPrices_FullMethodName = "/priceStream.PriceService/StreamPrices"
)

// PriceServiceClient is the client API for PriceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PriceServiceClient interface {
	StreamPrices(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[Request, ExtProductPrice], error)
}

type priceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPriceServiceClient(cc grpc.ClientConnInterface) PriceServiceClient {
	return &priceServiceClient{cc}
}

func (c *priceServiceClient) StreamPrices(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[Request, ExtProductPrice], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &PriceService_ServiceDesc.Streams[0], PriceService_StreamPrices_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[Request, ExtProductPrice]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PriceService_StreamPricesClient = grpc.BidiStreamingClient[Request, ExtProductPrice]

// PriceServiceServer is the server API for PriceService service.
// All implementations must embed UnimplementedPriceServiceServer
// for forward compatibility.
type PriceServiceServer interface {
	StreamPrices(grpc.BidiStreamingServer[Request, ExtProductPrice]) error
	mustEmbedUnimplementedPriceServiceServer()
}

// UnimplementedPriceServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPriceServiceServer struct{}

func (UnimplementedPriceServiceServer) StreamPrices(grpc.BidiStreamingServer[Request, ExtProductPrice]) error {
	return status.Errorf(codes.Unimplemented, "method StreamPrices not implemented")
}
func (UnimplementedPriceServiceServer) mustEmbedUnimplementedPriceServiceServer() {}
func (UnimplementedPriceServiceServer) testEmbeddedByValue()                      {}

// UnsafePriceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PriceServiceServer will
// result in compilation errors.
type UnsafePriceServiceServer interface {
	mustEmbedUnimplementedPriceServiceServer()
}

func RegisterPriceServiceServer(s grpc.ServiceRegistrar, srv PriceServiceServer) {
	// If the following call pancis, it indicates UnimplementedPriceServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PriceService_ServiceDesc, srv)
}

func _PriceService_StreamPrices_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PriceServiceServer).StreamPrices(&grpc.GenericServerStream[Request, ExtProductPrice]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PriceService_StreamPricesServer = grpc.BidiStreamingServer[Request, ExtProductPrice]

// PriceService_ServiceDesc is the grpc.ServiceDesc for PriceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PriceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "priceStream.PriceService",
	HandlerType: (*PriceServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamPrices",
			Handler:       _PriceService_StreamPrices_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protocpb/stream.proto",
}
