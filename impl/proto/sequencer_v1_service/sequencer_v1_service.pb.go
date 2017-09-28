// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sequencer_v1_service.proto

/*
Package sequencer_v1_service is a generated protocol buffer package.

It is generated from these files:
	sequencer_v1_service.proto

It has these top-level messages:
*/
package sequencer_v1_service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import keytransparency_v1_types "github.com/google/keytransparency/core/proto/keytransparency_v1_types"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SequencerService service

type SequencerServiceClient interface {
	// GetEpochs is a streaming API that sends epoch mutations upon creation.
	//
	// Returns the mutations of a newly created epoch.
	GetEpochs(ctx context.Context, in *keytransparency_v1_types.GetEpochsRequest, opts ...grpc.CallOption) (SequencerService_GetEpochsClient, error)
}

type sequencerServiceClient struct {
	cc *grpc.ClientConn
}

func NewSequencerServiceClient(cc *grpc.ClientConn) SequencerServiceClient {
	return &sequencerServiceClient{cc}
}

func (c *sequencerServiceClient) GetEpochs(ctx context.Context, in *keytransparency_v1_types.GetEpochsRequest, opts ...grpc.CallOption) (SequencerService_GetEpochsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SequencerService_serviceDesc.Streams[0], c.cc, "/sequencer.v1.service.SequencerService/GetEpochs", opts...)
	if err != nil {
		return nil, err
	}
	x := &sequencerServiceGetEpochsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SequencerService_GetEpochsClient interface {
	Recv() (*keytransparency_v1_types.GetEpochsResponse, error)
	grpc.ClientStream
}

type sequencerServiceGetEpochsClient struct {
	grpc.ClientStream
}

func (x *sequencerServiceGetEpochsClient) Recv() (*keytransparency_v1_types.GetEpochsResponse, error) {
	m := new(keytransparency_v1_types.GetEpochsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for SequencerService service

type SequencerServiceServer interface {
	// GetEpochs is a streaming API that sends epoch mutations upon creation.
	//
	// Returns the mutations of a newly created epoch.
	GetEpochs(*keytransparency_v1_types.GetEpochsRequest, SequencerService_GetEpochsServer) error
}

func RegisterSequencerServiceServer(s *grpc.Server, srv SequencerServiceServer) {
	s.RegisterService(&_SequencerService_serviceDesc, srv)
}

func _SequencerService_GetEpochs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(keytransparency_v1_types.GetEpochsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SequencerServiceServer).GetEpochs(m, &sequencerServiceGetEpochsServer{stream})
}

type SequencerService_GetEpochsServer interface {
	Send(*keytransparency_v1_types.GetEpochsResponse) error
	grpc.ServerStream
}

type sequencerServiceGetEpochsServer struct {
	grpc.ServerStream
}

func (x *sequencerServiceGetEpochsServer) Send(m *keytransparency_v1_types.GetEpochsResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _SequencerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sequencer.v1.service.SequencerService",
	HandlerType: (*SequencerServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetEpochs",
			Handler:       _SequencerService_GetEpochs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "sequencer_v1_service.proto",
}

func init() { proto.RegisterFile("sequencer_v1_service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 219 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x8f, 0xbf, 0x4a, 0xc5, 0x30,
	0x14, 0x87, 0xa9, 0x83, 0x60, 0x27, 0x8d, 0x2e, 0x06, 0x27, 0x47, 0x85, 0xc4, 0xea, 0xe6, 0x2e,
	0xee, 0xf6, 0x01, 0x4a, 0x1a, 0x0f, 0x6d, 0xd0, 0xe6, 0xc4, 0x9c, 0xd3, 0x40, 0x47, 0x7d, 0x05,
	0xc1, 0x17, 0xf3, 0x15, 0x7c, 0x10, 0xe9, 0x9f, 0xdb, 0xa1, 0x70, 0xe1, 0xce, 0xdf, 0xe1, 0xf7,
	0x7d, 0x27, 0x97, 0x04, 0x1f, 0x3d, 0x78, 0x0b, 0xb1, 0x4a, 0x45, 0x45, 0x10, 0x93, 0xb3, 0xa0,
	0x42, 0x44, 0x46, 0x71, 0xb1, 0x32, 0x95, 0x0a, 0xb5, 0x30, 0xf9, 0xda, 0x38, 0x6e, 0xfb, 0x5a,
	0x59, 0xec, 0x74, 0x83, 0xd8, 0xbc, 0x83, 0x7e, 0x83, 0x81, 0xa3, 0xf1, 0x14, 0x4c, 0x04, 0x6f,
	0x07, 0x6d, 0x31, 0x82, 0x9e, 0x36, 0xb6, 0x68, 0x94, 0xf0, 0x10, 0x80, 0xf6, 0x82, 0xd9, 0x2d,
	0xaf, 0x96, 0x69, 0x13, 0x9c, 0x36, 0xde, 0x23, 0x1b, 0x76, 0xe8, 0x17, 0x7a, 0xff, 0x93, 0xe5,
	0xa7, 0xe5, 0x2e, 0xae, 0x9c, 0xc3, 0xc4, 0x67, 0x96, 0x9f, 0x3c, 0x03, 0x3f, 0x05, 0xb4, 0x2d,
	0x89, 0x1b, 0xb5, 0x31, 0x8c, 0x3f, 0xcc, 0x86, 0xf5, 0xe8, 0x65, 0x9c, 0x20, 0x96, 0xb7, 0x07,
	0xdd, 0x52, 0x40, 0x4f, 0x70, 0x7d, 0xf9, 0xf5, 0xfb, 0xf7, 0x7d, 0x74, 0x2e, 0xce, 0x74, 0x2a,
	0x34, 0x4c, 0xec, 0x91, 0x38, 0x82, 0xe9, 0xee, 0xb2, 0xfa, 0x78, 0xea, 0x7b, 0xf8, 0x0f, 0x00,
	0x00, 0xff, 0xff, 0xaa, 0xe2, 0x26, 0xd2, 0x57, 0x01, 0x00, 0x00,
}
