// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: kwil/common/types.proto

package commonpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DataType int32

const (
	DataType_BEGIN_DATA_TYPES DataType = 0 // the first enum must be 0 in Proto3.  INVALID_TYPE is a valid Kwil enum, so I make this 0
	DataType_INVALID_TYPE     DataType = 100
	DataType_NULL             DataType = 101
	DataType_STRING           DataType = 102
	DataType_INT32            DataType = 103
	DataType_INT64            DataType = 104
	DataType_BOOLEAN          DataType = 105
)

// Enum value maps for DataType.
var (
	DataType_name = map[int32]string{
		0:   "BEGIN_DATA_TYPES",
		100: "INVALID_TYPE",
		101: "NULL",
		102: "STRING",
		103: "INT32",
		104: "INT64",
		105: "BOOLEAN",
	}
	DataType_value = map[string]int32{
		"BEGIN_DATA_TYPES": 0,
		"INVALID_TYPE":     100,
		"NULL":             101,
		"STRING":           102,
		"INT32":            103,
		"INT64":            104,
		"BOOLEAN":          105,
	}
)

func (x DataType) Enum() *DataType {
	p := new(DataType)
	*p = x
	return p
}

func (x DataType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DataType) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[0].Descriptor()
}

func (DataType) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[0]
}

func (x DataType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DataType.Descriptor instead.
func (DataType) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{0}
}

type AttributeType int32

const (
	AttributeType_BEGIN_ATTRIBUTE_TYPES AttributeType = 0
	AttributeType_INVALID_ATTRIBUTE     AttributeType = 100
	AttributeType_PRIMARY_KEY           AttributeType = 101
	AttributeType_UNIQUE                AttributeType = 102
	AttributeType_NOT_NULL              AttributeType = 103
	AttributeType_DEFAULT               AttributeType = 104
	AttributeType_MIN                   AttributeType = 105
	AttributeType_MAX                   AttributeType = 106
	AttributeType_MIN_LENGTH            AttributeType = 107
	AttributeType_MAX_LENGTH            AttributeType = 108
)

// Enum value maps for AttributeType.
var (
	AttributeType_name = map[int32]string{
		0:   "BEGIN_ATTRIBUTE_TYPES",
		100: "INVALID_ATTRIBUTE",
		101: "PRIMARY_KEY",
		102: "UNIQUE",
		103: "NOT_NULL",
		104: "DEFAULT",
		105: "MIN",
		106: "MAX",
		107: "MIN_LENGTH",
		108: "MAX_LENGTH",
	}
	AttributeType_value = map[string]int32{
		"BEGIN_ATTRIBUTE_TYPES": 0,
		"INVALID_ATTRIBUTE":     100,
		"PRIMARY_KEY":           101,
		"UNIQUE":                102,
		"NOT_NULL":              103,
		"DEFAULT":               104,
		"MIN":                   105,
		"MAX":                   106,
		"MIN_LENGTH":            107,
		"MAX_LENGTH":            108,
	}
)

func (x AttributeType) Enum() *AttributeType {
	p := new(AttributeType)
	*p = x
	return p
}

func (x AttributeType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AttributeType) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[1].Descriptor()
}

func (AttributeType) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[1]
}

func (x AttributeType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AttributeType.Descriptor instead.
func (AttributeType) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{1}
}

type IndexType int32

const (
	IndexType_BEGIN_INDEX_TYPES IndexType = 0
	IndexType_INVALID_INDEX     IndexType = 100
	IndexType_BTREE             IndexType = 101
)

// Enum value maps for IndexType.
var (
	IndexType_name = map[int32]string{
		0:   "BEGIN_INDEX_TYPES",
		100: "INVALID_INDEX",
		101: "BTREE",
	}
	IndexType_value = map[string]int32{
		"BEGIN_INDEX_TYPES": 0,
		"INVALID_INDEX":     100,
		"BTREE":             101,
	}
)

func (x IndexType) Enum() *IndexType {
	p := new(IndexType)
	*p = x
	return p
}

func (x IndexType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (IndexType) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[2].Descriptor()
}

func (IndexType) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[2]
}

func (x IndexType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use IndexType.Descriptor instead.
func (IndexType) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{2}
}

type ModifierType int32

const (
	ModifierType_NO_MODIFIER      ModifierType = 0 // since modifier can be left blank, the NO_MODIFIER value is 0
	ModifierType_INVALID_MODIFIER ModifierType = 100
	ModifierType_CALLER           ModifierType = 101
)

// Enum value maps for ModifierType.
var (
	ModifierType_name = map[int32]string{
		0:   "NO_MODIFIER",
		100: "INVALID_MODIFIER",
		101: "CALLER",
	}
	ModifierType_value = map[string]int32{
		"NO_MODIFIER":      0,
		"INVALID_MODIFIER": 100,
		"CALLER":           101,
	}
)

func (x ModifierType) Enum() *ModifierType {
	p := new(ModifierType)
	*p = x
	return p
}

func (x ModifierType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ModifierType) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[3].Descriptor()
}

func (ModifierType) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[3]
}

func (x ModifierType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ModifierType.Descriptor instead.
func (ModifierType) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{3}
}

type ComparisonOperator int32

const (
	ComparisonOperator_BEGIN_OPERATOR_TYPES           ComparisonOperator = 0
	ComparisonOperator_OPERATOR_INVALID               ComparisonOperator = 100
	ComparisonOperator_OPERATOR_EQUAL                 ComparisonOperator = 101
	ComparisonOperator_OPERATOR_NOT_EQUAL             ComparisonOperator = 102
	ComparisonOperator_OPERATOR_GREATER_THAN          ComparisonOperator = 103
	ComparisonOperator_OPERATOR_GREATER_THAN_OR_EQUAL ComparisonOperator = 104
	ComparisonOperator_OPERATOR_LESS_THAN             ComparisonOperator = 105
	ComparisonOperator_OPERATOR_LESS_THAN_OR_EQUAL    ComparisonOperator = 106
)

// Enum value maps for ComparisonOperator.
var (
	ComparisonOperator_name = map[int32]string{
		0:   "BEGIN_OPERATOR_TYPES",
		100: "OPERATOR_INVALID",
		101: "OPERATOR_EQUAL",
		102: "OPERATOR_NOT_EQUAL",
		103: "OPERATOR_GREATER_THAN",
		104: "OPERATOR_GREATER_THAN_OR_EQUAL",
		105: "OPERATOR_LESS_THAN",
		106: "OPERATOR_LESS_THAN_OR_EQUAL",
	}
	ComparisonOperator_value = map[string]int32{
		"BEGIN_OPERATOR_TYPES":           0,
		"OPERATOR_INVALID":               100,
		"OPERATOR_EQUAL":                 101,
		"OPERATOR_NOT_EQUAL":             102,
		"OPERATOR_GREATER_THAN":          103,
		"OPERATOR_GREATER_THAN_OR_EQUAL": 104,
		"OPERATOR_LESS_THAN":             105,
		"OPERATOR_LESS_THAN_OR_EQUAL":    106,
	}
)

func (x ComparisonOperator) Enum() *ComparisonOperator {
	p := new(ComparisonOperator)
	*p = x
	return p
}

func (x ComparisonOperator) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ComparisonOperator) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[4].Descriptor()
}

func (ComparisonOperator) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[4]
}

func (x ComparisonOperator) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ComparisonOperator.Descriptor instead.
func (ComparisonOperator) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{4}
}

type QueryType int32

const (
	QueryType_BEGIN_QUERY_TYPES QueryType = 0
	QueryType_QUERY_INVALID     QueryType = 100
	QueryType_QUERY_INSERT      QueryType = 101
	QueryType_QUERY_UPDATE      QueryType = 102
	QueryType_QUERY_DELETE      QueryType = 103
)

// Enum value maps for QueryType.
var (
	QueryType_name = map[int32]string{
		0:   "BEGIN_QUERY_TYPES",
		100: "QUERY_INVALID",
		101: "QUERY_INSERT",
		102: "QUERY_UPDATE",
		103: "QUERY_DELETE",
	}
	QueryType_value = map[string]int32{
		"BEGIN_QUERY_TYPES": 0,
		"QUERY_INVALID":     100,
		"QUERY_INSERT":      101,
		"QUERY_UPDATE":      102,
		"QUERY_DELETE":      103,
	}
)

func (x QueryType) Enum() *QueryType {
	p := new(QueryType)
	*p = x
	return p
}

func (x QueryType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryType) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_common_types_proto_enumTypes[5].Descriptor()
}

func (QueryType) Type() protoreflect.EnumType {
	return &file_kwil_common_types_proto_enumTypes[5]
}

func (x QueryType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryType.Descriptor instead.
func (QueryType) EnumDescriptor() ([]byte, []int) {
	return file_kwil_common_types_proto_rawDescGZIP(), []int{5}
}

var File_kwil_common_types_proto protoreflect.FileDescriptor

var file_kwil_common_types_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2a, 0x6b, 0x0a, 0x08, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a,
	0x10, 0x42, 0x45, 0x47, 0x49, 0x4e, 0x5f, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x53, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x54,
	0x59, 0x50, 0x45, 0x10, 0x64, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x55, 0x4c, 0x4c, 0x10, 0x65, 0x12,
	0x0a, 0x0a, 0x06, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x10, 0x66, 0x12, 0x09, 0x0a, 0x05, 0x49,
	0x4e, 0x54, 0x33, 0x32, 0x10, 0x67, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4e, 0x54, 0x36, 0x34, 0x10,
	0x68, 0x12, 0x0b, 0x0a, 0x07, 0x42, 0x4f, 0x4f, 0x4c, 0x45, 0x41, 0x4e, 0x10, 0x69, 0x2a, 0xab,
	0x01, 0x0a, 0x0d, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x19, 0x0a, 0x15, 0x42, 0x45, 0x47, 0x49, 0x4e, 0x5f, 0x41, 0x54, 0x54, 0x52, 0x49, 0x42,
	0x55, 0x54, 0x45, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x53, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x49,
	0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x41, 0x54, 0x54, 0x52, 0x49, 0x42, 0x55, 0x54, 0x45,
	0x10, 0x64, 0x12, 0x0f, 0x0a, 0x0b, 0x50, 0x52, 0x49, 0x4d, 0x41, 0x52, 0x59, 0x5f, 0x4b, 0x45,
	0x59, 0x10, 0x65, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x4e, 0x49, 0x51, 0x55, 0x45, 0x10, 0x66, 0x12,
	0x0c, 0x0a, 0x08, 0x4e, 0x4f, 0x54, 0x5f, 0x4e, 0x55, 0x4c, 0x4c, 0x10, 0x67, 0x12, 0x0b, 0x0a,
	0x07, 0x44, 0x45, 0x46, 0x41, 0x55, 0x4c, 0x54, 0x10, 0x68, 0x12, 0x07, 0x0a, 0x03, 0x4d, 0x49,
	0x4e, 0x10, 0x69, 0x12, 0x07, 0x0a, 0x03, 0x4d, 0x41, 0x58, 0x10, 0x6a, 0x12, 0x0e, 0x0a, 0x0a,
	0x4d, 0x49, 0x4e, 0x5f, 0x4c, 0x45, 0x4e, 0x47, 0x54, 0x48, 0x10, 0x6b, 0x12, 0x0e, 0x0a, 0x0a,
	0x4d, 0x41, 0x58, 0x5f, 0x4c, 0x45, 0x4e, 0x47, 0x54, 0x48, 0x10, 0x6c, 0x2a, 0x40, 0x0a, 0x09,
	0x49, 0x6e, 0x64, 0x65, 0x78, 0x54, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x42, 0x45, 0x47,
	0x49, 0x4e, 0x5f, 0x49, 0x4e, 0x44, 0x45, 0x58, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x53, 0x10, 0x00,
	0x12, 0x11, 0x0a, 0x0d, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x49, 0x4e, 0x44, 0x45,
	0x58, 0x10, 0x64, 0x12, 0x09, 0x0a, 0x05, 0x42, 0x54, 0x52, 0x45, 0x45, 0x10, 0x65, 0x2a, 0x41,
	0x0a, 0x0c, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0f,
	0x0a, 0x0b, 0x4e, 0x4f, 0x5f, 0x4d, 0x4f, 0x44, 0x49, 0x46, 0x49, 0x45, 0x52, 0x10, 0x00, 0x12,
	0x14, 0x0a, 0x10, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4d, 0x4f, 0x44, 0x49, 0x46,
	0x49, 0x45, 0x52, 0x10, 0x64, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x41, 0x4c, 0x4c, 0x45, 0x52, 0x10,
	0x65, 0x2a, 0xe8, 0x01, 0x0a, 0x12, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x72, 0x69, 0x73, 0x6f, 0x6e,
	0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x14, 0x42, 0x45, 0x47, 0x49,
	0x4e, 0x5f, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x53,
	0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x49,
	0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x64, 0x12, 0x12, 0x0a, 0x0e, 0x4f, 0x50, 0x45, 0x52,
	0x41, 0x54, 0x4f, 0x52, 0x5f, 0x45, 0x51, 0x55, 0x41, 0x4c, 0x10, 0x65, 0x12, 0x16, 0x0a, 0x12,
	0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x51, 0x55,
	0x41, 0x4c, 0x10, 0x66, 0x12, 0x19, 0x0a, 0x15, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52,
	0x5f, 0x47, 0x52, 0x45, 0x41, 0x54, 0x45, 0x52, 0x5f, 0x54, 0x48, 0x41, 0x4e, 0x10, 0x67, 0x12,
	0x22, 0x0a, 0x1e, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x47, 0x52, 0x45, 0x41,
	0x54, 0x45, 0x52, 0x5f, 0x54, 0x48, 0x41, 0x4e, 0x5f, 0x4f, 0x52, 0x5f, 0x45, 0x51, 0x55, 0x41,
	0x4c, 0x10, 0x68, 0x12, 0x16, 0x0a, 0x12, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f,
	0x4c, 0x45, 0x53, 0x53, 0x5f, 0x54, 0x48, 0x41, 0x4e, 0x10, 0x69, 0x12, 0x1f, 0x0a, 0x1b, 0x4f,
	0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x4c, 0x45, 0x53, 0x53, 0x5f, 0x54, 0x48, 0x41,
	0x4e, 0x5f, 0x4f, 0x52, 0x5f, 0x45, 0x51, 0x55, 0x41, 0x4c, 0x10, 0x6a, 0x2a, 0x6b, 0x0a, 0x09,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x42, 0x45, 0x47,
	0x49, 0x4e, 0x5f, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x53, 0x10, 0x00,
	0x12, 0x11, 0x0a, 0x0d, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49,
	0x44, 0x10, 0x64, 0x12, 0x10, 0x0a, 0x0c, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x49, 0x4e, 0x53,
	0x45, 0x52, 0x54, 0x10, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x55,
	0x50, 0x44, 0x41, 0x54, 0x45, 0x10, 0x66, 0x12, 0x10, 0x0a, 0x0c, 0x51, 0x55, 0x45, 0x52, 0x59,
	0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x67, 0x42, 0x17, 0x5a, 0x15, 0x6b, 0x77, 0x69,
	0x6c, 0x2f, 0x78, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kwil_common_types_proto_rawDescOnce sync.Once
	file_kwil_common_types_proto_rawDescData = file_kwil_common_types_proto_rawDesc
)

func file_kwil_common_types_proto_rawDescGZIP() []byte {
	file_kwil_common_types_proto_rawDescOnce.Do(func() {
		file_kwil_common_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_kwil_common_types_proto_rawDescData)
	})
	return file_kwil_common_types_proto_rawDescData
}

var file_kwil_common_types_proto_enumTypes = make([]protoimpl.EnumInfo, 6)
var file_kwil_common_types_proto_goTypes = []interface{}{
	(DataType)(0),           // 0: common.DataType
	(AttributeType)(0),      // 1: common.AttributeType
	(IndexType)(0),          // 2: common.IndexType
	(ModifierType)(0),       // 3: common.ModifierType
	(ComparisonOperator)(0), // 4: common.ComparisonOperator
	(QueryType)(0),          // 5: common.QueryType
}
var file_kwil_common_types_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kwil_common_types_proto_init() }
func file_kwil_common_types_proto_init() {
	if File_kwil_common_types_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kwil_common_types_proto_rawDesc,
			NumEnums:      6,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kwil_common_types_proto_goTypes,
		DependencyIndexes: file_kwil_common_types_proto_depIdxs,
		EnumInfos:         file_kwil_common_types_proto_enumTypes,
	}.Build()
	File_kwil_common_types_proto = out.File
	file_kwil_common_types_proto_rawDesc = nil
	file_kwil_common_types_proto_goTypes = nil
	file_kwil_common_types_proto_depIdxs = nil
}
