// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: kwil/tx/v1/service.proto

package txpb

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_kwil_tx_v1_service_proto protoreflect.FileDescriptor

var file_kwil_tx_v1_service_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x74, 0x78, 0x1a, 0x1a,
	0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x72, 0x6f, 0x61, 0x64,
	0x63, 0x61, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x6b, 0x77, 0x69, 0x6c,
	0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x16, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x6b, 0x77, 0x69, 0x6c,
	0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x18, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x6b, 0x77, 0x69,
	0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31,
	0x2f, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x6b, 0x77, 0x69,
	0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x6b, 0x77, 0x69, 0x6c,
	0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xe4, 0x09, 0x0a, 0x09, 0x54, 0x78, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x56, 0x0a, 0x09, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73,
	0x74, 0x12, 0x14, 0x2e, 0x74, 0x78, 0x2e, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x74, 0x78, 0x2e, 0x42, 0x72, 0x6f,
	0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1c,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x3a, 0x01, 0x2a, 0x22, 0x11, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x62, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x12, 0x67, 0x0a, 0x0d,
	0x45, 0x73, 0x74, 0x69, 0x6d, 0x61, 0x74, 0x65, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x18, 0x2e,
	0x74, 0x78, 0x2e, 0x45, 0x73, 0x74, 0x69, 0x6d, 0x61, 0x74, 0x65, 0x50, 0x72, 0x69, 0x63, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x74, 0x78, 0x2e, 0x45, 0x73, 0x74,
	0x69, 0x6d, 0x61, 0x74, 0x65, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x3a, 0x01, 0x2a, 0x22, 0x16, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x73, 0x74, 0x69, 0x6d, 0x61, 0x74, 0x65, 0x5f,
	0x70, 0x72, 0x69, 0x63, 0x65, 0x12, 0x46, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x10,
	0x2e, 0x74, 0x78, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x11, 0x2e, 0x74, 0x78, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x3a, 0x01, 0x2a, 0x22, 0x0d,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x5f, 0x0a,
	0x0a, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x15, 0x2e, 0x74, 0x78,
	0x2e, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x16, 0x2e, 0x74, 0x78, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x1c, 0x12, 0x1a, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x7d, 0x12, 0x3f,
	0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x0f, 0x2e, 0x74, 0x78, 0x2e, 0x50, 0x69, 0x6e, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x74, 0x78, 0x2e, 0x50, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x0e, 0x12, 0x0c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x69, 0x6e, 0x67, 0x12,
	0x50, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x14, 0x2e, 0x74,
	0x78, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x15, 0x2e, 0x74, 0x78, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x10, 0x12, 0x0e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x67, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73,
	0x65, 0x73, 0x12, 0x18, 0x2e, 0x74, 0x78, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x74,
	0x78, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x12,
	0x19, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x7b, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x7d,
	0x2f, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73, 0x12, 0x61, 0x0a, 0x09, 0x47, 0x65,
	0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x14, 0x2e, 0x74, 0x78, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x74, 0x78, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x27, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x21, 0x12, 0x1f, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73, 0x2f,
	0x7b, 0x64, 0x62, 0x69, 0x64, 0x7d, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x75, 0x0a,
	0x10, 0x41, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x12, 0x1c, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x41, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1d, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x41, 0x70,
	0x70, 0x72, 0x6f, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x24,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1e, 0x3a, 0x01, 0x2a, 0x22, 0x19, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x12, 0x67, 0x0a, 0x0d, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x4a, 0x6f, 0x69, 0x6e, 0x12, 0x18, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x19, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x4a, 0x6f,
	0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x1b, 0x3a, 0x01, 0x2a, 0x22, 0x16, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x6a, 0x6f, 0x69, 0x6e, 0x12, 0x6b, 0x0a,
	0x0e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x4c, 0x65, 0x61, 0x76, 0x65, 0x12,
	0x19, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x4c, 0x65,
	0x61, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x74, 0x78, 0x2e,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x4c, 0x65, 0x61, 0x76, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x3a, 0x01,
	0x2a, 0x22, 0x17, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x5f, 0x6c, 0x65, 0x61, 0x76, 0x65, 0x12, 0x7d, 0x0a, 0x13, 0x56, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x4a, 0x6f, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x1e, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x4a, 0x6f, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1f, 0x2e, 0x74, 0x78, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x4a, 0x6f, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x12, 0x1d, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x6a, 0x6f,
	0x69, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x42, 0x0a, 0x04, 0x43, 0x61, 0x6c,
	0x6c, 0x12, 0x0f, 0x2e, 0x74, 0x78, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x10, 0x2e, 0x74, 0x78, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x17, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x3a, 0x01, 0x2a, 0x22,
	0x0c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x42, 0x35, 0x5a,
	0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c,
	0x74, 0x65, 0x61, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c, 0x2d, 0x64, 0x62, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x3b,
	0x74, 0x78, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_kwil_tx_v1_service_proto_goTypes = []interface{}{
	(*BroadcastRequest)(nil),            // 0: tx.BroadcastRequest
	(*EstimatePriceRequest)(nil),        // 1: tx.EstimatePriceRequest
	(*QueryRequest)(nil),                // 2: tx.QueryRequest
	(*GetAccountRequest)(nil),           // 3: tx.GetAccountRequest
	(*PingRequest)(nil),                 // 4: tx.PingRequest
	(*GetConfigRequest)(nil),            // 5: tx.GetConfigRequest
	(*ListDatabasesRequest)(nil),        // 6: tx.ListDatabasesRequest
	(*GetSchemaRequest)(nil),            // 7: tx.GetSchemaRequest
	(*ValidatorApprovalRequest)(nil),    // 8: tx.ValidatorApprovalRequest
	(*ValidatorJoinRequest)(nil),        // 9: tx.ValidatorJoinRequest
	(*ValidatorLeaveRequest)(nil),       // 10: tx.ValidatorLeaveRequest
	(*ValidatorJoinStatusRequest)(nil),  // 11: tx.ValidatorJoinStatusRequest
	(*CallRequest)(nil),                 // 12: tx.CallRequest
	(*BroadcastResponse)(nil),           // 13: tx.BroadcastResponse
	(*EstimatePriceResponse)(nil),       // 14: tx.EstimatePriceResponse
	(*QueryResponse)(nil),               // 15: tx.QueryResponse
	(*GetAccountResponse)(nil),          // 16: tx.GetAccountResponse
	(*PingResponse)(nil),                // 17: tx.PingResponse
	(*GetConfigResponse)(nil),           // 18: tx.GetConfigResponse
	(*ListDatabasesResponse)(nil),       // 19: tx.ListDatabasesResponse
	(*GetSchemaResponse)(nil),           // 20: tx.GetSchemaResponse
	(*ValidatorApprovalResponse)(nil),   // 21: tx.ValidatorApprovalResponse
	(*ValidatorJoinResponse)(nil),       // 22: tx.ValidatorJoinResponse
	(*ValidatorLeaveResponse)(nil),      // 23: tx.ValidatorLeaveResponse
	(*ValidatorJoinStatusResponse)(nil), // 24: tx.ValidatorJoinStatusResponse
	(*CallResponse)(nil),                // 25: tx.CallResponse
}
var file_kwil_tx_v1_service_proto_depIdxs = []int32{
	0,  // 0: tx.TxService.Broadcast:input_type -> tx.BroadcastRequest
	1,  // 1: tx.TxService.EstimatePrice:input_type -> tx.EstimatePriceRequest
	2,  // 2: tx.TxService.Query:input_type -> tx.QueryRequest
	3,  // 3: tx.TxService.GetAccount:input_type -> tx.GetAccountRequest
	4,  // 4: tx.TxService.Ping:input_type -> tx.PingRequest
	5,  // 5: tx.TxService.GetConfig:input_type -> tx.GetConfigRequest
	6,  // 6: tx.TxService.ListDatabases:input_type -> tx.ListDatabasesRequest
	7,  // 7: tx.TxService.GetSchema:input_type -> tx.GetSchemaRequest
	8,  // 8: tx.TxService.ApproveValidator:input_type -> tx.ValidatorApprovalRequest
	9,  // 9: tx.TxService.ValidatorJoin:input_type -> tx.ValidatorJoinRequest
	10, // 10: tx.TxService.ValidatorLeave:input_type -> tx.ValidatorLeaveRequest
	11, // 11: tx.TxService.ValidatorJoinStatus:input_type -> tx.ValidatorJoinStatusRequest
	12, // 12: tx.TxService.Call:input_type -> tx.CallRequest
	13, // 13: tx.TxService.Broadcast:output_type -> tx.BroadcastResponse
	14, // 14: tx.TxService.EstimatePrice:output_type -> tx.EstimatePriceResponse
	15, // 15: tx.TxService.Query:output_type -> tx.QueryResponse
	16, // 16: tx.TxService.GetAccount:output_type -> tx.GetAccountResponse
	17, // 17: tx.TxService.Ping:output_type -> tx.PingResponse
	18, // 18: tx.TxService.GetConfig:output_type -> tx.GetConfigResponse
	19, // 19: tx.TxService.ListDatabases:output_type -> tx.ListDatabasesResponse
	20, // 20: tx.TxService.GetSchema:output_type -> tx.GetSchemaResponse
	21, // 21: tx.TxService.ApproveValidator:output_type -> tx.ValidatorApprovalResponse
	22, // 22: tx.TxService.ValidatorJoin:output_type -> tx.ValidatorJoinResponse
	23, // 23: tx.TxService.ValidatorLeave:output_type -> tx.ValidatorLeaveResponse
	24, // 24: tx.TxService.ValidatorJoinStatus:output_type -> tx.ValidatorJoinStatusResponse
	25, // 25: tx.TxService.Call:output_type -> tx.CallResponse
	13, // [13:26] is the sub-list for method output_type
	0,  // [0:13] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_kwil_tx_v1_service_proto_init() }
func file_kwil_tx_v1_service_proto_init() {
	if File_kwil_tx_v1_service_proto != nil {
		return
	}
	file_kwil_tx_v1_broadcast_proto_init()
	file_kwil_tx_v1_price_proto_init()
	file_kwil_tx_v1_query_proto_init()
	file_kwil_tx_v1_ping_proto_init()
	file_kwil_tx_v1_account_proto_init()
	file_kwil_tx_v1_config_proto_init()
	file_kwil_tx_v1_list_proto_init()
	file_kwil_tx_v1_dataset_proto_init()
	file_kwil_tx_v1_call_proto_init()
	file_kwil_tx_v1_validator_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kwil_tx_v1_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kwil_tx_v1_service_proto_goTypes,
		DependencyIndexes: file_kwil_tx_v1_service_proto_depIdxs,
	}.Build()
	File_kwil_tx_v1_service_proto = out.File
	file_kwil_tx_v1_service_proto_rawDesc = nil
	file_kwil_tx_v1_service_proto_goTypes = nil
	file_kwil_tx_v1_service_proto_depIdxs = nil
}
