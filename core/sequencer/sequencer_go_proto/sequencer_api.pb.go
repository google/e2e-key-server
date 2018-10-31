// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sequencer_api.proto

package sequencer_go_proto // import "github.com/google/keytransparency/core/sequencer/sequencer_go_proto"

/*
Key Transparency Sequencer

The Key Transparency Sequencer API supplies an api for applying mutations to the current
state of the map.
*/

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import empty "github.com/golang/protobuf/ptypes/empty"

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

type MapMetadata struct {
	// sources is a map from log source IDs to the (low, high] range of primary
	// keys in each slice that were used to construct this map revision.
	Sources              map[int64]*MapMetadata_SourceSlice `protobuf:"bytes,2,rep,name=sources,proto3" json:"sources,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}                           `json:"-"`
	XXX_unrecognized     []byte                             `json:"-"`
	XXX_sizecache        int32                              `json:"-"`
}

func (m *MapMetadata) Reset()         { *m = MapMetadata{} }
func (m *MapMetadata) String() string { return proto.CompactTextString(m) }
func (*MapMetadata) ProtoMessage()    {}
func (*MapMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_sequencer_api_6b093140ccdf94f7, []int{0}
}
func (m *MapMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MapMetadata.Unmarshal(m, b)
}
func (m *MapMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MapMetadata.Marshal(b, m, deterministic)
}
func (dst *MapMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MapMetadata.Merge(dst, src)
}
func (m *MapMetadata) XXX_Size() int {
	return xxx_messageInfo_MapMetadata.Size(m)
}
func (m *MapMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_MapMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_MapMetadata proto.InternalMessageInfo

func (m *MapMetadata) GetSources() map[int64]*MapMetadata_SourceSlice {
	if m != nil {
		return m.Sources
	}
	return nil
}

// SourceSlice is the range of inputs that have been included in a map
// revision.
type MapMetadata_SourceSlice struct {
	// lowest_watermark is the lowest primary key (exclusive) of the source
	// log that has been incorporated into this map revision. The primary
	// keys of logged items MUST be monotonically increasing.
	LowestWatermark int64 `protobuf:"varint,1,opt,name=lowest_watermark,json=lowestWatermark,proto3" json:"lowest_watermark,omitempty"`
	// highest_watermark is the highest primary key (inclusive) of the source
	// log that has been incorporated into this map revision. The primary keys
	// of logged items MUST be monotonically increasing.
	HighestWatermark     int64    `protobuf:"varint,2,opt,name=highest_watermark,json=highestWatermark,proto3" json:"highest_watermark,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MapMetadata_SourceSlice) Reset()         { *m = MapMetadata_SourceSlice{} }
func (m *MapMetadata_SourceSlice) String() string { return proto.CompactTextString(m) }
func (*MapMetadata_SourceSlice) ProtoMessage()    {}
func (*MapMetadata_SourceSlice) Descriptor() ([]byte, []int) {
	return fileDescriptor_sequencer_api_6b093140ccdf94f7, []int{0, 0}
}
func (m *MapMetadata_SourceSlice) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MapMetadata_SourceSlice.Unmarshal(m, b)
}
func (m *MapMetadata_SourceSlice) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MapMetadata_SourceSlice.Marshal(b, m, deterministic)
}
func (dst *MapMetadata_SourceSlice) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MapMetadata_SourceSlice.Merge(dst, src)
}
func (m *MapMetadata_SourceSlice) XXX_Size() int {
	return xxx_messageInfo_MapMetadata_SourceSlice.Size(m)
}
func (m *MapMetadata_SourceSlice) XXX_DiscardUnknown() {
	xxx_messageInfo_MapMetadata_SourceSlice.DiscardUnknown(m)
}

var xxx_messageInfo_MapMetadata_SourceSlice proto.InternalMessageInfo

func (m *MapMetadata_SourceSlice) GetLowestWatermark() int64 {
	if m != nil {
		return m.LowestWatermark
	}
	return 0
}

func (m *MapMetadata_SourceSlice) GetHighestWatermark() int64 {
	if m != nil {
		return m.HighestWatermark
	}
	return 0
}

// CreateEpochRequest contains information needed to create a new epoch.
type CreateEpochRequest struct {
	// directory_id is the directory to apply the mutations to.
	DirectoryId string `protobuf:"bytes,1,opt,name=directory_id,json=directoryId,proto3" json:"directory_id,omitempty"`
	// revision is the expected revision of the new epoch.
	Revision             int64    `protobuf:"varint,3,opt,name=revision,proto3" json:"revision,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateEpochRequest) Reset()         { *m = CreateEpochRequest{} }
func (m *CreateEpochRequest) String() string { return proto.CompactTextString(m) }
func (*CreateEpochRequest) ProtoMessage()    {}
func (*CreateEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_sequencer_api_6b093140ccdf94f7, []int{1}
}
func (m *CreateEpochRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateEpochRequest.Unmarshal(m, b)
}
func (m *CreateEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateEpochRequest.Marshal(b, m, deterministic)
}
func (dst *CreateEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateEpochRequest.Merge(dst, src)
}
func (m *CreateEpochRequest) XXX_Size() int {
	return xxx_messageInfo_CreateEpochRequest.Size(m)
}
func (m *CreateEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateEpochRequest proto.InternalMessageInfo

func (m *CreateEpochRequest) GetDirectoryId() string {
	if m != nil {
		return m.DirectoryId
	}
	return ""
}

func (m *CreateEpochRequest) GetRevision() int64 {
	if m != nil {
		return m.Revision
	}
	return 0
}

// RunBatchRequest triggers the sequencing of a batch of mutations for a
// directory, with the batch size governed by the request parameters.
type RunBatchRequest struct {
	// directory_id is the directory to run for.
	DirectoryId string `protobuf:"bytes,1,opt,name=directory_id,json=directoryId,proto3" json:"directory_id,omitempty"`
	// min_batch is the minimum number of items in a batch.
	// If less than min_batch items are available, nothing happens.
	// TODO(#1047): Replace with timeout so items in the log get processed
	// eventually.
	MinBatch int32 `protobuf:"varint,2,opt,name=min_batch,json=minBatch,proto3" json:"min_batch,omitempty"`
	// max_batch is the maximum number of items in a batch.
	MaxBatch             int32    `protobuf:"varint,3,opt,name=max_batch,json=maxBatch,proto3" json:"max_batch,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RunBatchRequest) Reset()         { *m = RunBatchRequest{} }
func (m *RunBatchRequest) String() string { return proto.CompactTextString(m) }
func (*RunBatchRequest) ProtoMessage()    {}
func (*RunBatchRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_sequencer_api_6b093140ccdf94f7, []int{2}
}
func (m *RunBatchRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RunBatchRequest.Unmarshal(m, b)
}
func (m *RunBatchRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RunBatchRequest.Marshal(b, m, deterministic)
}
func (dst *RunBatchRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RunBatchRequest.Merge(dst, src)
}
func (m *RunBatchRequest) XXX_Size() int {
	return xxx_messageInfo_RunBatchRequest.Size(m)
}
func (m *RunBatchRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RunBatchRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RunBatchRequest proto.InternalMessageInfo

func (m *RunBatchRequest) GetDirectoryId() string {
	if m != nil {
		return m.DirectoryId
	}
	return ""
}

func (m *RunBatchRequest) GetMinBatch() int32 {
	if m != nil {
		return m.MinBatch
	}
	return 0
}

func (m *RunBatchRequest) GetMaxBatch() int32 {
	if m != nil {
		return m.MaxBatch
	}
	return 0
}

// PublishBatchRequest copies all SignedMapHeads into the Log of SignedMapHeads.
type PublishBatchRequest struct {
	DirectoryId          string   `protobuf:"bytes,1,opt,name=directory_id,json=directoryId,proto3" json:"directory_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PublishBatchRequest) Reset()         { *m = PublishBatchRequest{} }
func (m *PublishBatchRequest) String() string { return proto.CompactTextString(m) }
func (*PublishBatchRequest) ProtoMessage()    {}
func (*PublishBatchRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_sequencer_api_6b093140ccdf94f7, []int{3}
}
func (m *PublishBatchRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PublishBatchRequest.Unmarshal(m, b)
}
func (m *PublishBatchRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PublishBatchRequest.Marshal(b, m, deterministic)
}
func (dst *PublishBatchRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PublishBatchRequest.Merge(dst, src)
}
func (m *PublishBatchRequest) XXX_Size() int {
	return xxx_messageInfo_PublishBatchRequest.Size(m)
}
func (m *PublishBatchRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PublishBatchRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PublishBatchRequest proto.InternalMessageInfo

func (m *PublishBatchRequest) GetDirectoryId() string {
	if m != nil {
		return m.DirectoryId
	}
	return ""
}

func init() {
	proto.RegisterType((*MapMetadata)(nil), "google.keytransparency.sequencer.MapMetadata")
	proto.RegisterMapType((map[int64]*MapMetadata_SourceSlice)(nil), "google.keytransparency.sequencer.MapMetadata.SourcesEntry")
	proto.RegisterType((*MapMetadata_SourceSlice)(nil), "google.keytransparency.sequencer.MapMetadata.SourceSlice")
	proto.RegisterType((*CreateEpochRequest)(nil), "google.keytransparency.sequencer.CreateEpochRequest")
	proto.RegisterType((*RunBatchRequest)(nil), "google.keytransparency.sequencer.RunBatchRequest")
	proto.RegisterType((*PublishBatchRequest)(nil), "google.keytransparency.sequencer.PublishBatchRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// KeyTransparencySequencerClient is the client API for KeyTransparencySequencer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KeyTransparencySequencerClient interface {
	// RunBatch reads outstanding mutations and calls CreateEpoch.
	RunBatch(ctx context.Context, in *RunBatchRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// CreateEpoch applies the contained mutations to the current map root.
	// If this method fails, it must be retried with the same arguments.
	CreateEpoch(ctx context.Context, in *CreateEpochRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// PublishBatch copies the MapRoots of all known map revisions into the Log
	// of MapRoots.
	PublishBatch(ctx context.Context, in *PublishBatchRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type keyTransparencySequencerClient struct {
	cc *grpc.ClientConn
}

func NewKeyTransparencySequencerClient(cc *grpc.ClientConn) KeyTransparencySequencerClient {
	return &keyTransparencySequencerClient{cc}
}

func (c *keyTransparencySequencerClient) RunBatch(ctx context.Context, in *RunBatchRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/google.keytransparency.sequencer.KeyTransparencySequencer/RunBatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencySequencerClient) CreateEpoch(ctx context.Context, in *CreateEpochRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/google.keytransparency.sequencer.KeyTransparencySequencer/CreateEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyTransparencySequencerClient) PublishBatch(ctx context.Context, in *PublishBatchRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/google.keytransparency.sequencer.KeyTransparencySequencer/PublishBatch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeyTransparencySequencerServer is the server API for KeyTransparencySequencer service.
type KeyTransparencySequencerServer interface {
	// RunBatch reads outstanding mutations and calls CreateEpoch.
	RunBatch(context.Context, *RunBatchRequest) (*empty.Empty, error)
	// CreateEpoch applies the contained mutations to the current map root.
	// If this method fails, it must be retried with the same arguments.
	CreateEpoch(context.Context, *CreateEpochRequest) (*empty.Empty, error)
	// PublishBatch copies the MapRoots of all known map revisions into the Log
	// of MapRoots.
	PublishBatch(context.Context, *PublishBatchRequest) (*empty.Empty, error)
}

func RegisterKeyTransparencySequencerServer(s *grpc.Server, srv KeyTransparencySequencerServer) {
	s.RegisterService(&_KeyTransparencySequencer_serviceDesc, srv)
}

func _KeyTransparencySequencer_RunBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencySequencerServer).RunBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.keytransparency.sequencer.KeyTransparencySequencer/RunBatch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencySequencerServer).RunBatch(ctx, req.(*RunBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencySequencer_CreateEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencySequencerServer).CreateEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.keytransparency.sequencer.KeyTransparencySequencer/CreateEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencySequencerServer).CreateEpoch(ctx, req.(*CreateEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyTransparencySequencer_PublishBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyTransparencySequencerServer).PublishBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.keytransparency.sequencer.KeyTransparencySequencer/PublishBatch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyTransparencySequencerServer).PublishBatch(ctx, req.(*PublishBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeyTransparencySequencer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.keytransparency.sequencer.KeyTransparencySequencer",
	HandlerType: (*KeyTransparencySequencerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RunBatch",
			Handler:    _KeyTransparencySequencer_RunBatch_Handler,
		},
		{
			MethodName: "CreateEpoch",
			Handler:    _KeyTransparencySequencer_CreateEpoch_Handler,
		},
		{
			MethodName: "PublishBatch",
			Handler:    _KeyTransparencySequencer_PublishBatch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sequencer_api.proto",
}

func init() { proto.RegisterFile("sequencer_api.proto", fileDescriptor_sequencer_api_6b093140ccdf94f7) }

var fileDescriptor_sequencer_api_6b093140ccdf94f7 = []byte{
	// 469 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0xc7, 0x69, 0xd2, 0x42, 0x76, 0x52, 0x69, 0xc1, 0x93, 0x50, 0x95, 0xdd, 0x94, 0x5c, 0x0d,
	0x21, 0x25, 0xa2, 0x80, 0x34, 0x76, 0xb9, 0xa9, 0x17, 0x14, 0x4d, 0xa0, 0x74, 0xd2, 0x24, 0x24,
	0x14, 0xb9, 0xe9, 0x21, 0xb1, 0x9a, 0xc4, 0xc1, 0x71, 0xb6, 0xe5, 0xb9, 0x78, 0x08, 0x5e, 0x0b,
	0xd5, 0xf9, 0x20, 0x8c, 0x8f, 0xaa, 0xbb, 0x8a, 0x73, 0xfe, 0xff, 0xf3, 0x3b, 0x3e, 0xc7, 0x36,
	0x1c, 0x15, 0xf8, 0xad, 0xc4, 0x2c, 0x44, 0x11, 0xd0, 0x9c, 0xb9, 0xb9, 0xe0, 0x92, 0x93, 0x69,
	0xc4, 0x79, 0x94, 0xa0, 0xbb, 0xc1, 0x4a, 0x0a, 0x9a, 0x15, 0x39, 0x15, 0x98, 0x85, 0x95, 0xdb,
	0x79, 0xed, 0xe3, 0xda, 0xe1, 0x29, 0xff, 0xaa, 0xfc, 0xea, 0x61, 0x9a, 0xcb, 0xaa, 0x4e, 0x77,
	0x7e, 0x68, 0x60, 0x5e, 0xd2, 0xfc, 0x12, 0x25, 0x5d, 0x53, 0x49, 0xc9, 0x15, 0x3c, 0x29, 0x78,
	0x29, 0x42, 0x2c, 0x26, 0xda, 0x54, 0x3f, 0x31, 0x67, 0x67, 0xee, 0xae, 0x02, 0x6e, 0x2f, 0xdf,
	0x5d, 0xd6, 0xc9, 0xf3, 0x4c, 0x8a, 0xca, 0x6f, 0x51, 0x36, 0x82, 0x59, 0x0b, 0xcb, 0x84, 0x85,
	0x48, 0x5e, 0x80, 0x95, 0xf0, 0x5b, 0x2c, 0x64, 0x70, 0x4b, 0x25, 0x8a, 0x94, 0x8a, 0xcd, 0x64,
	0x30, 0x1d, 0x9c, 0xe8, 0xfe, 0x61, 0x1d, 0xbf, 0x6e, 0xc3, 0xe4, 0x25, 0x3c, 0x8d, 0x59, 0x14,
	0xff, 0xee, 0xd5, 0x94, 0xd7, 0x6a, 0x84, 0xce, 0x6c, 0x97, 0x30, 0xee, 0xd7, 0x27, 0x16, 0xe8,
	0x1b, 0xac, 0x1a, 0xf4, 0x76, 0x49, 0x3e, 0xc2, 0xe8, 0x86, 0x26, 0x25, 0x2a, 0x84, 0x39, 0x7b,
	0xf7, 0x90, 0xe6, 0x54, 0x0f, 0x7e, 0xcd, 0x39, 0xd3, 0x4e, 0x07, 0x8b, 0xa1, 0x31, 0xb0, 0x34,
	0x27, 0x00, 0x72, 0x21, 0x90, 0x4a, 0x9c, 0xe7, 0x3c, 0x8c, 0xfd, 0x2d, 0xa0, 0x90, 0xe4, 0x39,
	0x8c, 0xd7, 0x4c, 0x60, 0x28, 0xb9, 0xa8, 0x02, 0xb6, 0x56, 0x7b, 0x39, 0xf0, 0xcd, 0x2e, 0xf6,
	0x7e, 0x4d, 0x6c, 0x30, 0x04, 0xde, 0xb0, 0x82, 0xf1, 0x6c, 0xa2, 0xab, 0xad, 0x76, 0xff, 0x8b,
	0xa1, 0xa1, 0x59, 0xfa, 0x62, 0x68, 0x0c, 0xad, 0x91, 0x93, 0xc1, 0xa1, 0x5f, 0x66, 0xe7, 0x54,
	0xee, 0x45, 0x3f, 0x86, 0x83, 0x94, 0x65, 0xc1, 0x6a, 0x9b, 0xa6, 0xba, 0x1e, 0xf9, 0x46, 0xca,
	0x6a, 0x8c, 0x12, 0xe9, 0x5d, 0x23, 0xea, 0x8d, 0x48, 0xef, 0x94, 0xe8, 0x9c, 0xc2, 0xd1, 0xa7,
	0x72, 0x95, 0xb0, 0x22, 0xde, 0xb3, 0xe6, 0xec, 0xbb, 0x06, 0x93, 0x0f, 0x58, 0x5d, 0xf5, 0x26,
	0xba, 0x6c, 0x07, 0x4a, 0xae, 0xc1, 0x68, 0xdb, 0x20, 0xaf, 0x76, 0xcf, 0xff, 0x5e, 0xcb, 0xf6,
	0xb3, 0x36, 0xa5, 0xbd, 0xce, 0xee, 0x7c, 0x7b, 0x9d, 0x9d, 0x47, 0xe4, 0x0b, 0x98, 0xbd, 0x03,
	0x20, 0x6f, 0x76, 0xb3, 0xff, 0x3c, 0xaf, 0xff, 0xe0, 0x03, 0x18, 0xf7, 0xc7, 0x41, 0xde, 0xee,
	0xe6, 0xff, 0x65, 0x7c, 0xff, 0x2e, 0x70, 0x3e, 0xff, 0x7c, 0x11, 0x31, 0x19, 0x97, 0x2b, 0x37,
	0xe4, 0xa9, 0xd7, 0x3c, 0xda, 0x7b, 0x70, 0x2f, 0xe4, 0x02, 0xbd, 0xae, 0xc2, 0xaf, 0x55, 0x10,
	0xf1, 0xa0, 0x26, 0x3e, 0x56, 0x9f, 0xd7, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xff, 0xb6, 0xe9,
	0xbd, 0x2e, 0x04, 0x00, 0x00,
}
