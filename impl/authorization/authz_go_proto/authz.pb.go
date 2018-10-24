// Code generated by protoc-gen-go. DO NOT EDIT.
// source: authz.proto

package authz_go_proto // import "github.com/google/keytransparency/impl/authorization/authz_go_proto"

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

// AuthorizationPolicy contains an authorization policy.
type AuthorizationPolicy struct {
	// roles is a map of roles keyed by labels used in RoleLabels.
	Roles map[string]*AuthorizationPolicy_Role `protobuf:"bytes,2,rep,name=roles" json:"roles,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// resource_to_role_labels specifies the authorization policy keyed by resource
	// map_id|app_id concatenation as a string.
	ResourceToRoleLabels map[string]*AuthorizationPolicy_RoleLabels `protobuf:"bytes,3,rep,name=resource_to_role_labels,json=resourceToRoleLabels" json:"resource_to_role_labels,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	XXX_NoUnkeyedLiteral struct{}                                   `json:"-"`
	XXX_unrecognized     []byte                                     `json:"-"`
	XXX_sizecache        int32                                      `json:"-"`
}

func (m *AuthorizationPolicy) Reset()         { *m = AuthorizationPolicy{} }
func (m *AuthorizationPolicy) String() string { return proto.CompactTextString(m) }
func (*AuthorizationPolicy) ProtoMessage()    {}
func (*AuthorizationPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_authz_ff29cd762dfc07be, []int{0}
}
func (m *AuthorizationPolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthorizationPolicy.Unmarshal(m, b)
}
func (m *AuthorizationPolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthorizationPolicy.Marshal(b, m, deterministic)
}
func (dst *AuthorizationPolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthorizationPolicy.Merge(dst, src)
}
func (m *AuthorizationPolicy) XXX_Size() int {
	return xxx_messageInfo_AuthorizationPolicy.Size(m)
}
func (m *AuthorizationPolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthorizationPolicy.DiscardUnknown(m)
}

var xxx_messageInfo_AuthorizationPolicy proto.InternalMessageInfo

func (m *AuthorizationPolicy) GetRoles() map[string]*AuthorizationPolicy_Role {
	if m != nil {
		return m.Roles
	}
	return nil
}

func (m *AuthorizationPolicy) GetResourceToRoleLabels() map[string]*AuthorizationPolicy_RoleLabels {
	if m != nil {
		return m.ResourceToRoleLabels
	}
	return nil
}

// Resource contains the resource being accessed.
type AuthorizationPolicy_Resource struct {
	// directory_id contains the Key Transparency directory of this entry.
	DirectoryId string `protobuf:"bytes,1,opt,name=directory_id,json=directoryId" json:"directory_id,omitempty"`
	// app_id contains the application identity of this entry.
	AppId                string   `protobuf:"bytes,2,opt,name=app_id,json=appId" json:"app_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AuthorizationPolicy_Resource) Reset()         { *m = AuthorizationPolicy_Resource{} }
func (m *AuthorizationPolicy_Resource) String() string { return proto.CompactTextString(m) }
func (*AuthorizationPolicy_Resource) ProtoMessage()    {}
func (*AuthorizationPolicy_Resource) Descriptor() ([]byte, []int) {
	return fileDescriptor_authz_ff29cd762dfc07be, []int{0, 0}
}
func (m *AuthorizationPolicy_Resource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthorizationPolicy_Resource.Unmarshal(m, b)
}
func (m *AuthorizationPolicy_Resource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthorizationPolicy_Resource.Marshal(b, m, deterministic)
}
func (dst *AuthorizationPolicy_Resource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthorizationPolicy_Resource.Merge(dst, src)
}
func (m *AuthorizationPolicy_Resource) XXX_Size() int {
	return xxx_messageInfo_AuthorizationPolicy_Resource.Size(m)
}
func (m *AuthorizationPolicy_Resource) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthorizationPolicy_Resource.DiscardUnknown(m)
}

var xxx_messageInfo_AuthorizationPolicy_Resource proto.InternalMessageInfo

func (m *AuthorizationPolicy_Resource) GetDirectoryId() string {
	if m != nil {
		return m.DirectoryId
	}
	return ""
}

func (m *AuthorizationPolicy_Resource) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

// Role contains a specific identity of an authorization entry.
type AuthorizationPolicy_Role struct {
	// principals contains an application specific identifier for this entry.
	Principals           []string `protobuf:"bytes,1,rep,name=principals" json:"principals,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AuthorizationPolicy_Role) Reset()         { *m = AuthorizationPolicy_Role{} }
func (m *AuthorizationPolicy_Role) String() string { return proto.CompactTextString(m) }
func (*AuthorizationPolicy_Role) ProtoMessage()    {}
func (*AuthorizationPolicy_Role) Descriptor() ([]byte, []int) {
	return fileDescriptor_authz_ff29cd762dfc07be, []int{0, 1}
}
func (m *AuthorizationPolicy_Role) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthorizationPolicy_Role.Unmarshal(m, b)
}
func (m *AuthorizationPolicy_Role) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthorizationPolicy_Role.Marshal(b, m, deterministic)
}
func (dst *AuthorizationPolicy_Role) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthorizationPolicy_Role.Merge(dst, src)
}
func (m *AuthorizationPolicy_Role) XXX_Size() int {
	return xxx_messageInfo_AuthorizationPolicy_Role.Size(m)
}
func (m *AuthorizationPolicy_Role) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthorizationPolicy_Role.DiscardUnknown(m)
}

var xxx_messageInfo_AuthorizationPolicy_Role proto.InternalMessageInfo

func (m *AuthorizationPolicy_Role) GetPrincipals() []string {
	if m != nil {
		return m.Principals
	}
	return nil
}

// RoleLabels contains a lot of role labels identifying each role.
type AuthorizationPolicy_RoleLabels struct {
	Labels               []string `protobuf:"bytes,1,rep,name=labels" json:"labels,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AuthorizationPolicy_RoleLabels) Reset()         { *m = AuthorizationPolicy_RoleLabels{} }
func (m *AuthorizationPolicy_RoleLabels) String() string { return proto.CompactTextString(m) }
func (*AuthorizationPolicy_RoleLabels) ProtoMessage()    {}
func (*AuthorizationPolicy_RoleLabels) Descriptor() ([]byte, []int) {
	return fileDescriptor_authz_ff29cd762dfc07be, []int{0, 2}
}
func (m *AuthorizationPolicy_RoleLabels) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthorizationPolicy_RoleLabels.Unmarshal(m, b)
}
func (m *AuthorizationPolicy_RoleLabels) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthorizationPolicy_RoleLabels.Marshal(b, m, deterministic)
}
func (dst *AuthorizationPolicy_RoleLabels) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthorizationPolicy_RoleLabels.Merge(dst, src)
}
func (m *AuthorizationPolicy_RoleLabels) XXX_Size() int {
	return xxx_messageInfo_AuthorizationPolicy_RoleLabels.Size(m)
}
func (m *AuthorizationPolicy_RoleLabels) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthorizationPolicy_RoleLabels.DiscardUnknown(m)
}

var xxx_messageInfo_AuthorizationPolicy_RoleLabels proto.InternalMessageInfo

func (m *AuthorizationPolicy_RoleLabels) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func init() {
	proto.RegisterType((*AuthorizationPolicy)(nil), "google.keytransparency.impl.AuthorizationPolicy")
	proto.RegisterMapType((map[string]*AuthorizationPolicy_RoleLabels)(nil), "google.keytransparency.impl.AuthorizationPolicy.ResourceToRoleLabelsEntry")
	proto.RegisterMapType((map[string]*AuthorizationPolicy_Role)(nil), "google.keytransparency.impl.AuthorizationPolicy.RolesEntry")
	proto.RegisterType((*AuthorizationPolicy_Resource)(nil), "google.keytransparency.impl.AuthorizationPolicy.Resource")
	proto.RegisterType((*AuthorizationPolicy_Role)(nil), "google.keytransparency.impl.AuthorizationPolicy.Role")
	proto.RegisterType((*AuthorizationPolicy_RoleLabels)(nil), "google.keytransparency.impl.AuthorizationPolicy.RoleLabels")
}

func init() { proto.RegisterFile("authz.proto", fileDescriptor_authz_ff29cd762dfc07be) }

var fileDescriptor_authz_ff29cd762dfc07be = []byte{
	// 359 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xcf, 0x4a, 0xeb, 0x40,
	0x14, 0xc6, 0x49, 0xd3, 0x96, 0xf6, 0x74, 0x73, 0x99, 0x7b, 0xaf, 0xc6, 0x14, 0xa4, 0x88, 0x48,
	0x57, 0x09, 0x54, 0x04, 0x51, 0x10, 0x54, 0xba, 0xa8, 0xba, 0xb0, 0xc1, 0x95, 0x9b, 0x30, 0x4d,
	0x86, 0x74, 0xe8, 0x34, 0x67, 0x98, 0x4c, 0x84, 0x74, 0x27, 0xf8, 0xc4, 0x3e, 0x81, 0x24, 0x13,
	0xb1, 0x4a, 0x15, 0xec, 0x2a, 0x39, 0xff, 0x7e, 0xdf, 0x37, 0x33, 0x07, 0x7a, 0x34, 0xd7, 0xf3,
	0x95, 0x27, 0x15, 0x6a, 0x24, 0xfd, 0x04, 0x31, 0x11, 0xcc, 0x5b, 0xb0, 0x42, 0x2b, 0x9a, 0x66,
	0x92, 0x2a, 0x96, 0x46, 0x85, 0xc7, 0x97, 0x52, 0x1c, 0xbc, 0x36, 0xe1, 0xef, 0x65, 0xae, 0xe7,
	0xa8, 0xf8, 0x8a, 0x6a, 0x8e, 0xe9, 0x3d, 0x0a, 0x1e, 0x15, 0x64, 0x0a, 0x2d, 0x85, 0x82, 0x65,
	0x4e, 0x63, 0x60, 0x0f, 0x7b, 0xa3, 0x73, 0xef, 0x07, 0x88, 0xb7, 0x01, 0xe0, 0x05, 0xe5, 0xf4,
	0x38, 0xd5, 0xaa, 0x08, 0x0c, 0x89, 0x3c, 0x5b, 0xb0, 0xab, 0x58, 0x86, 0xb9, 0x8a, 0x58, 0xa8,
	0x31, 0x2c, 0xb3, 0xa1, 0xa0, 0x33, 0x26, 0x32, 0xc7, 0xae, 0x54, 0x6e, 0x7e, 0xaf, 0x52, 0xf3,
	0x1e, 0xb0, 0xd4, 0xbb, 0xab, 0x60, 0x46, 0xf4, 0x9f, 0xda, 0x50, 0x72, 0x2f, 0xa0, 0xf3, 0x3e,
	0x42, 0xfa, 0xd0, 0x8d, 0x71, 0x49, 0x79, 0x1a, 0xf2, 0xd8, 0xb1, 0x06, 0xd6, 0xb0, 0x1b, 0x74,
	0x4c, 0x62, 0x12, 0x93, 0xff, 0xd0, 0xa6, 0x52, 0x96, 0x95, 0x46, 0x55, 0x69, 0x51, 0x29, 0x27,
	0xb1, 0x7b, 0x04, 0xcd, 0x92, 0x46, 0xf6, 0x01, 0xa4, 0xe2, 0x69, 0xc4, 0x25, 0x15, 0x99, 0x63,
	0x0d, 0xec, 0x61, 0x37, 0x58, 0xcb, 0xb8, 0x87, 0x00, 0x1f, 0xaa, 0x64, 0x07, 0xda, 0xf5, 0x39,
	0x4d, 0x67, 0x1d, 0xb9, 0x68, 0xba, 0x8c, 0x63, 0xf2, 0x07, 0xec, 0x05, 0x2b, 0x6a, 0x27, 0xe5,
	0x2f, 0xb9, 0x85, 0xd6, 0x13, 0x15, 0x39, 0xab, 0x3c, 0xf4, 0x46, 0x27, 0x5b, 0x3d, 0x42, 0x60,
	0x18, 0x67, 0x8d, 0x53, 0xcb, 0x7d, 0xb1, 0x60, 0xef, 0xdb, 0x2b, 0xdb, 0x60, 0x60, 0xfa, 0xd9,
	0xc0, 0x76, 0x5b, 0x60, 0x24, 0xd6, 0x6c, 0x5c, 0x8d, 0x1f, 0xaf, 0x13, 0xae, 0xe7, 0xf9, 0xcc,
	0x8b, 0x70, 0xe9, 0x1b, 0xa6, 0xff, 0x85, 0xe9, 0x97, 0x4c, 0x9f, 0xae, 0x33, 0xab, 0x68, 0x15,
	0x26, 0x18, 0x56, 0x8b, 0x3d, 0x6b, 0x57, 0x9f, 0xe3, 0xb7, 0x00, 0x00, 0x00, 0xff, 0xff, 0xab,
	0x65, 0xd8, 0x7f, 0xee, 0x02, 0x00, 0x00,
}
