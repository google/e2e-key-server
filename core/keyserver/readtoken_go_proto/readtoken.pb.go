// Code generated by protoc-gen-go. DO NOT EDIT.
// source: readtoken.proto

package readtoken_go_proto // import "github.com/google/keytransparency/core/keyserver/readtoken_go_proto"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// ReadToken can be serialized and handed to users for pagination.
type ReadToken struct {
	// shard_id identifies the source for reading.
	ShardId int64 `protobuf:"varint,1,opt,name=shard_id,json=shardId,proto3" json:"shard_id,omitempty"`
	// low_watemark identifies the lowest (exclusive) row to return.
	LowWatermark         int64    `protobuf:"varint,2,opt,name=low_watermark,json=lowWatermark,proto3" json:"low_watermark,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReadToken) Reset()         { *m = ReadToken{} }
func (m *ReadToken) String() string { return proto.CompactTextString(m) }
func (*ReadToken) ProtoMessage()    {}
func (*ReadToken) Descriptor() ([]byte, []int) {
	return fileDescriptor_readtoken_451ac17fca853c14, []int{0}
}
func (m *ReadToken) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReadToken.Unmarshal(m, b)
}
func (m *ReadToken) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReadToken.Marshal(b, m, deterministic)
}
func (dst *ReadToken) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReadToken.Merge(dst, src)
}
func (m *ReadToken) XXX_Size() int {
	return xxx_messageInfo_ReadToken.Size(m)
}
func (m *ReadToken) XXX_DiscardUnknown() {
	xxx_messageInfo_ReadToken.DiscardUnknown(m)
}

var xxx_messageInfo_ReadToken proto.InternalMessageInfo

func (m *ReadToken) GetShardId() int64 {
	if m != nil {
		return m.ShardId
	}
	return 0
}

func (m *ReadToken) GetLowWatermark() int64 {
	if m != nil {
		return m.LowWatermark
	}
	return 0
}

func init() {
	proto.RegisterType((*ReadToken)(nil), "google.keytransparency.v1.ReadToken")
}

func init() { proto.RegisterFile("readtoken.proto", fileDescriptor_readtoken_451ac17fca853c14) }

var fileDescriptor_readtoken_451ac17fca853c14 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x4a, 0x4d, 0x4c,
	0x29, 0xc9, 0xcf, 0x4e, 0xcd, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x4c, 0xcf, 0xcf,
	0x4f, 0xcf, 0x49, 0xd5, 0xcb, 0x4e, 0xad, 0x2c, 0x29, 0x4a, 0xcc, 0x2b, 0x2e, 0x48, 0x2c, 0x4a,
	0xcd, 0x4b, 0xae, 0xd4, 0x2b, 0x33, 0x54, 0xf2, 0xe6, 0xe2, 0x0c, 0x4a, 0x4d, 0x4c, 0x09, 0x01,
	0xa9, 0x16, 0x92, 0xe4, 0xe2, 0x28, 0xce, 0x48, 0x2c, 0x4a, 0x89, 0xcf, 0x4c, 0x91, 0x60, 0x54,
	0x60, 0xd4, 0x60, 0x0e, 0x62, 0x07, 0xf3, 0x3d, 0x53, 0x84, 0x94, 0xb9, 0x78, 0x73, 0xf2, 0xcb,
	0xe3, 0xcb, 0x13, 0x4b, 0x52, 0x8b, 0x72, 0x13, 0x8b, 0xb2, 0x25, 0x98, 0xc0, 0xf2, 0x3c, 0x39,
	0xf9, 0xe5, 0xe1, 0x30, 0x31, 0x27, 0xd7, 0x28, 0xe7, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd,
	0xe4, 0xfc, 0x5c, 0x7d, 0x88, 0xa5, 0xfa, 0x68, 0x96, 0xea, 0x27, 0xe7, 0x17, 0x81, 0x05, 0x8b,
	0x53, 0x8b, 0xca, 0x52, 0x8b, 0xf4, 0xe1, 0x6e, 0x8d, 0x4f, 0xcf, 0x8f, 0x07, 0x3b, 0x37, 0x89,
	0x0d, 0x4c, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xcb, 0xf6, 0x33, 0x26, 0xc8, 0x00, 0x00,
	0x00,
}
