// Code generated by protoc-gen-go. DO NOT EDIT.
// source: monitor/v1/monitor.proto

// Monitor Service
//
// The Key Transparency monitor server service consists of APIs to fetch
// monitor results queried using the mutations API.

package monitor_go_proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	trillian "github.com/google/trillian"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	status "google.golang.org/genproto/googleapis/rpc/status"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status1 "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// GetStateRequest requests the verification state of a keytransparency
// directory for a particular point in time.
type GetStateRequest struct {
	// kt_url is the URL of the keytransparency server for which the monitoring
	// result will be returned.
	KtUrl string `protobuf:"bytes,2,opt,name=kt_url,json=ktUrl,proto3" json:"kt_url,omitempty"`
	// directory_id identifies the merkle tree being monitored.
	DirectoryId string `protobuf:"bytes,3,opt,name=directory_id,json=directoryId,proto3" json:"directory_id,omitempty"`
	// revision specifies the revision for which the monitoring results will
	// be returned (revisions start at 0).
	Revision             int64    `protobuf:"varint,1,opt,name=revision,proto3" json:"revision,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetStateRequest) Reset()         { *m = GetStateRequest{} }
func (m *GetStateRequest) String() string { return proto.CompactTextString(m) }
func (*GetStateRequest) ProtoMessage()    {}
func (*GetStateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_6c9cdd4901f6b9a2, []int{0}
}

func (m *GetStateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetStateRequest.Unmarshal(m, b)
}
func (m *GetStateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetStateRequest.Marshal(b, m, deterministic)
}
func (m *GetStateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetStateRequest.Merge(m, src)
}
func (m *GetStateRequest) XXX_Size() int {
	return xxx_messageInfo_GetStateRequest.Size(m)
}
func (m *GetStateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetStateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetStateRequest proto.InternalMessageInfo

func (m *GetStateRequest) GetKtUrl() string {
	if m != nil {
		return m.KtUrl
	}
	return ""
}

func (m *GetStateRequest) GetDirectoryId() string {
	if m != nil {
		return m.DirectoryId
	}
	return ""
}

func (m *GetStateRequest) GetRevision() int64 {
	if m != nil {
		return m.Revision
	}
	return 0
}

// State represents the monitor's evaluation of a Key Transparency directory
// at a particular revision.
type State struct {
	// smr contains the map root for the sparse Merkle Tree signed with the
	// monitor's key on success. If the checks were not successful the
	// smr will be empty. The revisions are encoded into the smr map_revision.
	Smr *trillian.SignedMapRoot `protobuf:"bytes,1,opt,name=smr,proto3" json:"smr,omitempty"`
	// seen_time contains the time when this particular signed map root was
	// retrieved and processed.
	SeenTime *timestamp.Timestamp `protobuf:"bytes,2,opt,name=seen_time,json=seenTime,proto3" json:"seen_time,omitempty"`
	// errors contains a list of errors representing the verification checks
	// that failed while monitoring the key-transparency server.
	Errors               []*status.Status `protobuf:"bytes,3,rep,name=errors,proto3" json:"errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_6c9cdd4901f6b9a2, []int{1}
}

func (m *State) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_State.Unmarshal(m, b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_State.Marshal(b, m, deterministic)
}
func (m *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(m, src)
}
func (m *State) XXX_Size() int {
	return xxx_messageInfo_State.Size(m)
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetSmr() *trillian.SignedMapRoot {
	if m != nil {
		return m.Smr
	}
	return nil
}

func (m *State) GetSeenTime() *timestamp.Timestamp {
	if m != nil {
		return m.SeenTime
	}
	return nil
}

func (m *State) GetErrors() []*status.Status {
	if m != nil {
		return m.Errors
	}
	return nil
}

func init() {
	proto.RegisterType((*GetStateRequest)(nil), "google.keytransparency.monitor.v1.GetStateRequest")
	proto.RegisterType((*State)(nil), "google.keytransparency.monitor.v1.State")
}

func init() { proto.RegisterFile("monitor/v1/monitor.proto", fileDescriptor_6c9cdd4901f6b9a2) }

var fileDescriptor_6c9cdd4901f6b9a2 = []byte{
	// 453 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x93, 0xc1, 0x8a, 0xd4, 0x4c,
	0x10, 0xc7, 0xc9, 0x86, 0x9d, 0x6f, 0xb6, 0xe7, 0x43, 0xa1, 0x41, 0x36, 0x04, 0xc1, 0xd9, 0x39,
	0x8d, 0x1e, 0xba, 0xd9, 0x78, 0x10, 0x3c, 0x2e, 0xe8, 0xba, 0xc8, 0x1c, 0xcc, 0xe8, 0xc5, 0x4b,
	0xe8, 0xcd, 0x94, 0xb1, 0x99, 0xa4, 0x3b, 0x56, 0x57, 0x06, 0x86, 0x61, 0x2e, 0x5e, 0x3d, 0x7a,
	0xf0, 0x51, 0x7c, 0x02, 0x9f, 0xc0, 0x57, 0xf0, 0x41, 0x24, 0x9d, 0x64, 0x58, 0xd6, 0x83, 0x82,
	0x78, 0x49, 0x52, 0xe9, 0x7f, 0x25, 0xf5, 0xff, 0x55, 0x15, 0x8b, 0x2a, 0x6b, 0x34, 0x59, 0x94,
	0x9b, 0x73, 0xd9, 0x3f, 0x8a, 0x1a, 0x2d, 0x59, 0x7e, 0x56, 0x58, 0x5b, 0x94, 0x20, 0xd6, 0xb0,
	0x25, 0x54, 0xc6, 0xd5, 0x0a, 0xc1, 0xe4, 0x5b, 0x31, 0xa8, 0x36, 0xe7, 0xf1, 0xfd, 0x4e, 0x22,
	0x55, 0xad, 0xa5, 0x32, 0xc6, 0x92, 0x22, 0x6d, 0x8d, 0xeb, 0x3e, 0x10, 0x3f, 0xe8, 0x4f, 0x7d,
	0x74, 0xdd, 0xbc, 0x93, 0xa4, 0x2b, 0x70, 0xa4, 0xaa, 0xba, 0x17, 0x9c, 0xf6, 0x02, 0xac, 0x73,
	0xe9, 0x48, 0x51, 0x33, 0x64, 0xde, 0x21, 0xd4, 0x65, 0xa9, 0x95, 0xe9, 0xe2, 0x59, 0xc1, 0xee,
	0x5e, 0x02, 0x2d, 0x49, 0x11, 0xa4, 0xf0, 0xa1, 0x01, 0x47, 0xfc, 0x1e, 0x1b, 0xad, 0x29, 0x6b,
	0xb0, 0x8c, 0x8e, 0xa6, 0xc1, 0xfc, 0x24, 0x3d, 0x5e, 0xd3, 0x1b, 0x2c, 0xf9, 0x19, 0xfb, 0x7f,
	0xa5, 0x11, 0x72, 0xb2, 0xb8, 0xcd, 0xf4, 0x2a, 0x0a, 0xfd, 0xe1, 0xe4, 0xf0, 0xee, 0x6a, 0xc5,
	0x63, 0x36, 0x46, 0xd8, 0x68, 0xa7, 0xad, 0x89, 0x82, 0x69, 0x30, 0x0f, 0xd3, 0x43, 0x3c, 0xfb,
	0x12, 0xb0, 0x63, 0xff, 0x1b, 0xfe, 0x90, 0x85, 0xae, 0x42, 0x2f, 0x98, 0x24, 0xa7, 0xe2, 0x50,
	0xd0, 0x52, 0x17, 0x06, 0x56, 0x0b, 0x55, 0xa7, 0xd6, 0x52, 0xda, 0x6a, 0xf8, 0x13, 0x76, 0xe2,
	0x00, 0x4c, 0xd6, 0xda, 0xf3, 0xd5, 0x4c, 0x92, 0x58, 0xf4, 0xf0, 0x06, 0xef, 0xe2, 0xf5, 0xe0,
	0x3d, 0x1d, 0xb7, 0xe2, 0x36, 0xe4, 0x8f, 0xd8, 0x08, 0x10, 0x2d, 0xba, 0x28, 0x9c, 0x86, 0xf3,
	0x49, 0xc2, 0x87, 0x2c, 0xac, 0x73, 0xb1, 0xf4, 0x40, 0xd2, 0x5e, 0x91, 0x7c, 0x0a, 0xd9, 0x7f,
	0x8b, 0x8e, 0x3c, 0xff, 0x1a, 0xb0, 0xf1, 0xc0, 0x83, 0x27, 0xe2, 0xb7, 0x7d, 0x12, 0xb7, 0xe0,
	0xc5, 0xf3, 0x3f, 0xc8, 0xf1, 0x09, 0xb3, 0xc5, 0xc7, 0xef, 0x3f, 0x3e, 0x1f, 0x5d, 0xf2, 0x67,
	0xf2, 0xc6, 0x9c, 0x38, 0xc0, 0x0d, 0xa0, 0x93, 0xbb, 0xae, 0x03, 0x7b, 0x39, 0xe0, 0xd5, 0xe0,
	0xe4, 0xee, 0x26, 0xff, 0xbd, 0xef, 0x2b, 0xb8, 0xa7, 0x65, 0x7b, 0x25, 0xfe, 0x2d, 0x60, 0x7c,
	0x28, 0xe6, 0x62, 0x9b, 0xf6, 0xd8, 0xff, 0xb1, 0x87, 0x57, 0xde, 0xc3, 0x4b, 0x7e, 0xf5, 0x77,
	0x1e, 0xe4, 0x6e, 0x18, 0x93, 0xfd, 0xc5, 0x8b, 0xb7, 0xcf, 0x0b, 0x4d, 0xef, 0x9b, 0x6b, 0x91,
	0xdb, 0x4a, 0xf6, 0x63, 0x7c, 0xab, 0x10, 0x99, 0x5b, 0xec, 0x56, 0xe3, 0xd7, 0x15, 0xcb, 0x0a,
	0x9b, 0x75, 0xa3, 0x31, 0xf2, 0xb7, 0xc7, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xdb, 0xae, 0xd1,
	0xd7, 0x88, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MonitorClient is the client API for Monitor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MonitorClient interface {
	// GetSignedMapRoot returns the latest valid signed map root the monitor
	// observed. Additionally, the response contains extra data necessary to
	// reproduce errors on failure.
	//
	// Returns the signed map root for the latest revision the monitor observed. If
	// the monitor could not reconstruct the map root given the set of mutations
	// from the previous to the current revision it won't sign the map root and
	// additional data will be provided to reproduce the failure.
	GetState(ctx context.Context, in *GetStateRequest, opts ...grpc.CallOption) (*State, error)
	// GetSignedMapRootByRevision returns the monitor's result for a specific map
	// revision.
	//
	// Returns the signed map root for the specified revision the monitor observed.
	// If the monitor could not reconstruct the map root given the set of
	// mutations from the previous to the current revision it won't sign the map
	// root and additional data will be provided to reproduce the failure.
	GetStateByRevision(ctx context.Context, in *GetStateRequest, opts ...grpc.CallOption) (*State, error)
}

type monitorClient struct {
	cc *grpc.ClientConn
}

func NewMonitorClient(cc *grpc.ClientConn) MonitorClient {
	return &monitorClient{cc}
}

func (c *monitorClient) GetState(ctx context.Context, in *GetStateRequest, opts ...grpc.CallOption) (*State, error) {
	out := new(State)
	err := c.cc.Invoke(ctx, "/google.keytransparency.monitor.v1.Monitor/GetState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) GetStateByRevision(ctx context.Context, in *GetStateRequest, opts ...grpc.CallOption) (*State, error) {
	out := new(State)
	err := c.cc.Invoke(ctx, "/google.keytransparency.monitor.v1.Monitor/GetStateByRevision", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorServer is the server API for Monitor service.
type MonitorServer interface {
	// GetSignedMapRoot returns the latest valid signed map root the monitor
	// observed. Additionally, the response contains extra data necessary to
	// reproduce errors on failure.
	//
	// Returns the signed map root for the latest revision the monitor observed. If
	// the monitor could not reconstruct the map root given the set of mutations
	// from the previous to the current revision it won't sign the map root and
	// additional data will be provided to reproduce the failure.
	GetState(context.Context, *GetStateRequest) (*State, error)
	// GetSignedMapRootByRevision returns the monitor's result for a specific map
	// revision.
	//
	// Returns the signed map root for the specified revision the monitor observed.
	// If the monitor could not reconstruct the map root given the set of
	// mutations from the previous to the current revision it won't sign the map
	// root and additional data will be provided to reproduce the failure.
	GetStateByRevision(context.Context, *GetStateRequest) (*State, error)
}

// UnimplementedMonitorServer can be embedded to have forward compatible implementations.
type UnimplementedMonitorServer struct {
}

func (*UnimplementedMonitorServer) GetState(ctx context.Context, req *GetStateRequest) (*State, error) {
	return nil, status1.Errorf(codes.Unimplemented, "method GetState not implemented")
}
func (*UnimplementedMonitorServer) GetStateByRevision(ctx context.Context, req *GetStateRequest) (*State, error) {
	return nil, status1.Errorf(codes.Unimplemented, "method GetStateByRevision not implemented")
}

func RegisterMonitorServer(s *grpc.Server, srv MonitorServer) {
	s.RegisterService(&_Monitor_serviceDesc, srv)
}

func _Monitor_GetState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).GetState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.keytransparency.monitor.v1.Monitor/GetState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).GetState(ctx, req.(*GetStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_GetStateByRevision_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).GetStateByRevision(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.keytransparency.monitor.v1.Monitor/GetStateByRevision",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).GetStateByRevision(ctx, req.(*GetStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Monitor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.keytransparency.monitor.v1.Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetState",
			Handler:    _Monitor_GetState_Handler,
		},
		{
			MethodName: "GetStateByRevision",
			Handler:    _Monitor_GetStateByRevision_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "monitor/v1/monitor.proto",
}
