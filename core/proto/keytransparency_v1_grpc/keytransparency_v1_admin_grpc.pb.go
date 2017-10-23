// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/keytransparency_v1_grpc/keytransparency_v1_admin_grpc.proto

package keytransparency_v1_grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import google_protobuf4 "github.com/golang/protobuf/ptypes/empty"
import keytransparency_v1_proto1 "github.com/google/keytransparency/core/proto/keytransparency_v1_proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for KeyTransparencyAdminService service

type KeyTransparencyAdminServiceClient interface {
	// BatchUpdateEntries uses an authorized_public key to perform a set request on multiple entries at once.
	BatchUpdateEntries(ctx context.Context, in *keytransparency_v1_proto1.BatchUpdateEntriesRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.BatchUpdateEntriesResponse, error)
	// ListDomains returns a list of all domains this Key Transparency server
	// operates on.
	ListDomains(ctx context.Context, in *keytransparency_v1_proto1.ListDomainsRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.ListDomainsResponse, error)
	// GetDomain returns the confiuration information for a given domain.
	GetDomain(ctx context.Context, in *keytransparency_v1_proto1.GetDomainRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.GetDomainResponse, error)
	// CreateDomain creates a new Trillian log/map pair.  A unique domainId must
	// be provided.  To create a new domain with the same name as a previously
	// deleted domain, a user must wait X days until the domain is garbage
	// collected.
	CreateDomain(ctx context.Context, in *keytransparency_v1_proto1.CreateDomainRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.CreateDomainResponse, error)
	// DeleteDomain marks a domain as deleted.  Domains will be garbage collected
	// after X days.
	DeleteDomain(ctx context.Context, in *keytransparency_v1_proto1.DeleteDomainRequest, opts ...grpc.CallOption) (*google_protobuf4.Empty, error)
	// UndeleteDomain marks a previously deleted domain as active if it has not
	// already been garbage collected.
	UndeleteDomain(ctx context.Context, in *keytransparency_v1_proto1.UndeleteDomainRequest, opts ...grpc.CallOption) (*google_protobuf4.Empty, error)
}

type keyTransparencyAdminServiceClient struct {
	cc *grpc.ClientConn
}

func NewKeyTransparencyAdminServiceClient(cc *grpc.ClientConn) KeyTransparencyAdminServiceClient {
	return &keyTransparencyAdminServiceClient{cc}
}

func (c *keyTransparencyAdminServiceClient) BatchUpdateEntries(ctx context.Context, in *keytransparency_v1_proto1.BatchUpdateEntriesRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.BatchUpdateEntriesResponse, error) {
	out := new(keytransparency_v1_proto1.BatchUpdateEntriesResponse)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/BatchUpdateEntries", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencyAdminServiceClient) ListDomains(ctx context.Context, in *keytransparency_v1_proto1.ListDomainsRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.ListDomainsResponse, error) {
	out := new(keytransparency_v1_proto1.ListDomainsResponse)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/ListDomains", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencyAdminServiceClient) GetDomain(ctx context.Context, in *keytransparency_v1_proto1.GetDomainRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.GetDomainResponse, error) {
	out := new(keytransparency_v1_proto1.GetDomainResponse)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/GetDomain", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencyAdminServiceClient) CreateDomain(ctx context.Context, in *keytransparency_v1_proto1.CreateDomainRequest, opts ...grpc.CallOption) (*keytransparency_v1_proto1.CreateDomainResponse, error) {
	out := new(keytransparency_v1_proto1.CreateDomainResponse)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/CreateDomain", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencyAdminServiceClient) DeleteDomain(ctx context.Context, in *keytransparency_v1_proto1.DeleteDomainRequest, opts ...grpc.CallOption) (*google_protobuf4.Empty, error) {
	out := new(google_protobuf4.Empty)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/DeleteDomain", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencyAdminServiceClient) UndeleteDomain(ctx context.Context, in *keytransparency_v1_proto1.UndeleteDomainRequest, opts ...grpc.CallOption) (*google_protobuf4.Empty, error) {
	out := new(google_protobuf4.Empty)
	err := grpc.Invoke(ctx, "/keytransparency.v1.grpc.KeyTransparencyAdminService/UndeleteDomain", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for KeyTransparencyAdminService service

type KeyTransparencyAdminServiceServer interface {
	// BatchUpdateEntries uses an authorized_public key to perform a set request on multiple entries at once.
	BatchUpdateEntries(context.Context, *keytransparency_v1_proto1.BatchUpdateEntriesRequest) (*keytransparency_v1_proto1.BatchUpdateEntriesResponse, error)
	// ListDomains returns a list of all domains this Key Transparency server
	// operates on.
	ListDomains(context.Context, *keytransparency_v1_proto1.ListDomainsRequest) (*keytransparency_v1_proto1.ListDomainsResponse, error)
	// GetDomain returns the confiuration information for a given domain.
	GetDomain(context.Context, *keytransparency_v1_proto1.GetDomainRequest) (*keytransparency_v1_proto1.GetDomainResponse, error)
	// CreateDomain creates a new Trillian log/map pair.  A unique domainId must
	// be provided.  To create a new domain with the same name as a previously
	// deleted domain, a user must wait X days until the domain is garbage
	// collected.
	CreateDomain(context.Context, *keytransparency_v1_proto1.CreateDomainRequest) (*keytransparency_v1_proto1.CreateDomainResponse, error)
	// DeleteDomain marks a domain as deleted.  Domains will be garbage collected
	// after X days.
	DeleteDomain(context.Context, *keytransparency_v1_proto1.DeleteDomainRequest) (*google_protobuf4.Empty, error)
	// UndeleteDomain marks a previously deleted domain as active if it has not
	// already been garbage collected.
	UndeleteDomain(context.Context, *keytransparency_v1_proto1.UndeleteDomainRequest) (*google_protobuf4.Empty, error)
}

func RegisterKeyTransparencyAdminServiceServer(s *grpc.Server, srv KeyTransparencyAdminServiceServer) {
	s.RegisterService(&_KeyTransparencyAdminService_serviceDesc, srv)
}

func _KeyTransparencyAdminService_BatchUpdateEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.BatchUpdateEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).BatchUpdateEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/BatchUpdateEntries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).BatchUpdateEntries(ctx, req.(*keytransparency_v1_proto1.BatchUpdateEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencyAdminService_ListDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.ListDomainsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).ListDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/ListDomains",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).ListDomains(ctx, req.(*keytransparency_v1_proto1.ListDomainsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencyAdminService_GetDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.GetDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).GetDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/GetDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).GetDomain(ctx, req.(*keytransparency_v1_proto1.GetDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencyAdminService_CreateDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.CreateDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).CreateDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/CreateDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).CreateDomain(ctx, req.(*keytransparency_v1_proto1.CreateDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencyAdminService_DeleteDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.DeleteDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).DeleteDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/DeleteDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).DeleteDomain(ctx, req.(*keytransparency_v1_proto1.DeleteDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencyAdminService_UndeleteDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(keytransparency_v1_proto1.UndeleteDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencyAdminServiceServer).UndeleteDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keytransparency.v1.grpc.KeyTransparencyAdminService/UndeleteDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencyAdminServiceServer).UndeleteDomain(ctx, req.(*keytransparency_v1_proto1.UndeleteDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeyTransparencyAdminService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "keytransparency.v1.grpc.KeyTransparencyAdminService",
	HandlerType: (*KeyTransparencyAdminServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BatchUpdateEntries",
			Handler:    _KeyTransparencyAdminService_BatchUpdateEntries_Handler,
		},
		{
			MethodName: "ListDomains",
			Handler:    _KeyTransparencyAdminService_ListDomains_Handler,
		},
		{
			MethodName: "GetDomain",
			Handler:    _KeyTransparencyAdminService_GetDomain_Handler,
		},
		{
			MethodName: "CreateDomain",
			Handler:    _KeyTransparencyAdminService_CreateDomain_Handler,
		},
		{
			MethodName: "DeleteDomain",
			Handler:    _KeyTransparencyAdminService_DeleteDomain_Handler,
		},
		{
			MethodName: "UndeleteDomain",
			Handler:    _KeyTransparencyAdminService_UndeleteDomain_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/keytransparency_v1_grpc/keytransparency_v1_admin_grpc.proto",
}

func init() {
	proto.RegisterFile("proto/keytransparency_v1_grpc/keytransparency_v1_admin_grpc.proto", fileDescriptor1)
}

var fileDescriptor1 = []byte{
	// 417 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0xbd, 0xce, 0xd3, 0x30,
	0x14, 0x86, 0x09, 0xc3, 0x27, 0x91, 0x56, 0xfc, 0x78, 0x68, 0x45, 0x8b, 0x04, 0xca, 0x84, 0x02,
	0xd8, 0x2a, 0x65, 0xea, 0xd6, 0x3f, 0x18, 0x60, 0x02, 0xba, 0xb0, 0x54, 0x4e, 0x72, 0x48, 0x2d,
	0x1a, 0x3b, 0xd8, 0x4e, 0xa4, 0x08, 0xb1, 0x80, 0xb8, 0x02, 0x16, 0x2e, 0x84, 0x3b, 0xe1, 0x16,
	0xb8, 0x10, 0x14, 0xc7, 0x41, 0xa1, 0xd4, 0x90, 0x6f, 0x4b, 0xf2, 0x3e, 0xe7, 0x9c, 0xe7, 0x48,
	0x76, 0xfc, 0x65, 0x2e, 0x85, 0x16, 0xe4, 0x1d, 0x54, 0x5a, 0x52, 0xae, 0x72, 0x2a, 0x81, 0xc7,
	0xd5, 0xbe, 0x9c, 0xed, 0x53, 0x99, 0xc7, 0xe7, 0xbe, 0xd3, 0x24, 0x63, 0xdc, 0xa4, 0xd8, 0xd4,
	0xa2, 0xf1, 0x09, 0x84, 0xcb, 0x19, 0xae, 0xe3, 0xc9, 0x9d, 0x54, 0x88, 0xf4, 0x08, 0x84, 0xe6,
	0x8c, 0x50, 0xce, 0x85, 0xa6, 0x9a, 0x09, 0xae, 0x9a, 0xb2, 0xc9, 0xd4, 0xa6, 0xe6, 0x2d, 0x2a,
	0xde, 0x12, 0xc8, 0x72, 0x5d, 0xd9, 0x70, 0xed, 0xd4, 0x72, 0x06, 0x8d, 0x97, 0x89, 0x9b, 0x26,
	0x8f, 0xbf, 0x5f, 0xf8, 0xd3, 0xe7, 0x50, 0xbd, 0xee, 0x80, 0xcb, 0x1a, 0x7a, 0x05, 0xb2, 0x64,
	0x31, 0xa0, 0x6f, 0x9e, 0x8f, 0x56, 0x54, 0xc7, 0x87, 0x5d, 0x9e, 0x50, 0x0d, 0x5b, 0xae, 0x25,
	0x03, 0x85, 0xe6, 0xf8, 0xcc, 0x42, 0x4d, 0xdf, 0xbf, 0xe9, 0x97, 0xf0, 0xbe, 0x00, 0xa5, 0x27,
	0x4f, 0x2e, 0x57, 0xa4, 0x72, 0xc1, 0x15, 0x04, 0xe3, 0x4f, 0x3f, 0x7e, 0x7e, 0xbd, 0x7a, 0x2b,
	0xb8, 0x41, 0xca, 0x19, 0x29, 0x14, 0x48, 0xb5, 0x88, 0x6a, 0x1a, 0x1d, 0xfd, 0xc1, 0x0b, 0xa6,
	0xf4, 0x46, 0x64, 0x94, 0x71, 0x85, 0x1e, 0xba, 0xbb, 0x77, 0xb0, 0xd6, 0xe5, 0x51, 0x4f, 0xda,
	0x4a, 0x5c, 0x41, 0x5f, 0x3c, 0xff, 0xda, 0x33, 0xb0, 0x01, 0x0a, 0xdd, 0xe5, 0xbf, 0xa1, 0x76,
	0xd4, 0x83, 0x5e, 0xac, 0x1d, 0x74, 0xd7, 0x6c, 0x7b, 0x1b, 0x8d, 0xeb, 0x6d, 0x93, 0xc6, 0x82,
	0x7c, 0x68, 0x1e, 0xf6, 0x2c, 0xf9, 0x58, 0x7b, 0x0c, 0xd7, 0x12, 0xa8, 0x06, 0xab, 0xf2, 0x8f,
	0x4d, 0xba, 0x5c, 0x6b, 0x83, 0xfb, 0xe2, 0x56, 0x68, 0x64, 0x84, 0x6e, 0x06, 0x83, 0x8e, 0xd0,
	0xc2, 0x0b, 0x51, 0xe9, 0x0f, 0x37, 0x70, 0x84, 0x3e, 0x1a, 0x5d, 0xae, 0xd5, 0x18, 0xe1, 0xe6,
	0x68, 0xe3, 0xf6, 0x68, 0xe3, 0x6d, 0x7d, 0xb4, 0xdb, 0xfd, 0x43, 0xe7, 0xfe, 0x9f, 0x3d, 0xff,
	0xfa, 0x8e, 0x27, 0xdd, 0xd1, 0xc4, 0x3d, 0xfa, 0x4f, 0xf2, 0x7f, 0xc3, 0xef, 0x9b, 0xe1, 0x41,
	0x78, 0xcf, 0x31, 0x7c, 0x51, 0xd8, 0x76, 0xab, 0xa7, 0x6f, 0x36, 0x29, 0xd3, 0x87, 0x22, 0xc2,
	0xb1, 0xc8, 0x88, 0xbd, 0xa5, 0x27, 0x16, 0x24, 0x16, 0xd2, 0x5e, 0x5d, 0xd7, 0xbf, 0x23, 0xba,
	0x30, 0xf1, 0xfc, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0x63, 0x8d, 0xf7, 0x92, 0x63, 0x04, 0x00,
	0x00,
}
