// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: kwil/tx/v1/tx.proto

package txpb

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

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Body          *Transaction_Body `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	Signature     *Signature        `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	Sender        []byte            `protobuf:"bytes,3,opt,name=sender,proto3" json:"sender,omitempty"`
	Serialization string            `protobuf:"bytes,4,opt,name=serialization,proto3" json:"serialization,omitempty"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_proto_rawDescGZIP(), []int{0}
}

func (x *Transaction) GetBody() *Transaction_Body {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Transaction) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Transaction) GetSender() []byte {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (x *Transaction) GetSerialization() string {
	if x != nil {
		return x.Serialization
	}
	return ""
}

type Signature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SignatureBytes []byte `protobuf:"bytes,1,opt,name=signature_bytes,proto3" json:"signature_bytes,omitempty"`
	SignatureType  string `protobuf:"bytes,2,opt,name=signature_type,proto3" json:"signature_type,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Signature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Signature) ProtoMessage() {}

func (x *Signature) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Signature.ProtoReflect.Descriptor instead.
func (*Signature) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_proto_rawDescGZIP(), []int{1}
}

func (x *Signature) GetSignatureBytes() []byte {
	if x != nil {
		return x.SignatureBytes
	}
	return nil
}

func (x *Signature) GetSignatureType() string {
	if x != nil {
		return x.SignatureType
	}
	return ""
}

type TransactionResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code      uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Log       string   `protobuf:"bytes,2,opt,name=log,proto3" json:"log,omitempty"`
	GasUsed   int64    `protobuf:"varint,3,opt,name=gas_used,proto3" json:"gas_used,omitempty"`
	GasWanted int64    `protobuf:"varint,4,opt,name=gas_wanted,proto3" json:"gas_wanted,omitempty"`
	Data      []byte   `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"` // Data contains the output of the transaction.
	Events    [][]byte `protobuf:"bytes,6,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *TransactionResult) Reset() {
	*x = TransactionResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionResult) ProtoMessage() {}

func (x *TransactionResult) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionResult.ProtoReflect.Descriptor instead.
func (*TransactionResult) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_proto_rawDescGZIP(), []int{2}
}

func (x *TransactionResult) GetCode() uint32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *TransactionResult) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

func (x *TransactionResult) GetGasUsed() int64 {
	if x != nil {
		return x.GasUsed
	}
	return 0
}

func (x *TransactionResult) GetGasWanted() int64 {
	if x != nil {
		return x.GasWanted
	}
	return 0
}

func (x *TransactionResult) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *TransactionResult) GetEvents() [][]byte {
	if x != nil {
		return x.Events
	}
	return nil
}

// deprecated
type TransactionStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     []byte   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Fee    string   `protobuf:"bytes,2,opt,name=fee,proto3" json:"fee,omitempty"`
	Status string   `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Errors []string `protobuf:"bytes,4,rep,name=errors,proto3" json:"errors,omitempty"`
}

func (x *TransactionStatus) Reset() {
	*x = TransactionStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionStatus) ProtoMessage() {}

func (x *TransactionStatus) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionStatus.ProtoReflect.Descriptor instead.
func (*TransactionStatus) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_proto_rawDescGZIP(), []int{3}
}

func (x *TransactionStatus) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *TransactionStatus) GetFee() string {
	if x != nil {
		return x.Fee
	}
	return ""
}

func (x *TransactionStatus) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *TransactionStatus) GetErrors() []string {
	if x != nil {
		return x.Errors
	}
	return nil
}

type Transaction_Body struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload     []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	PayloadType string `protobuf:"bytes,2,opt,name=payload_type,proto3" json:"payload_type,omitempty"`
	Fee         string `protobuf:"bytes,3,opt,name=fee,proto3" json:"fee,omitempty"`
	Nonce       uint64 `protobuf:"varint,4,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Description string `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *Transaction_Body) Reset() {
	*x = Transaction_Body{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction_Body) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction_Body) ProtoMessage() {}

func (x *Transaction_Body) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction_Body.ProtoReflect.Descriptor instead.
func (*Transaction_Body) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Transaction_Body) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Transaction_Body) GetPayloadType() string {
	if x != nil {
		return x.PayloadType
	}
	return ""
}

func (x *Transaction_Body) GetFee() string {
	if x != nil {
		return x.Fee
	}
	return ""
}

func (x *Transaction_Body) GetNonce() uint64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *Transaction_Body) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

var File_kwil_tx_v1_tx_proto protoreflect.FileDescriptor

var file_kwil_tx_v1_tx_proto_rawDesc = []byte{
	0x0a, 0x13, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x78, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x74, 0x78, 0x22, 0xb3, 0x02, 0x0a, 0x0b, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x28, 0x0a, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x74, 0x78, 0x2e, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x42, 0x6f, 0x64, 0x79, 0x52, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x12, 0x2b, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x74, 0x78, 0x2e, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x65, 0x72, 0x69,
	0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x8e,
	0x01, 0x0a, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x65, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x66, 0x65, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22,
	0x5d, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x28, 0x0a, 0x0f,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x22, 0xa1,
	0x01, 0x0a, 0x11, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x67, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6c, 0x6f, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x67, 0x61,
	0x73, 0x5f, 0x75, 0x73, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x67, 0x61,
	0x73, 0x5f, 0x75, 0x73, 0x65, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x67, 0x61, 0x73, 0x5f, 0x77, 0x61,
	0x6e, 0x74, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x67, 0x61, 0x73, 0x5f,
	0x77, 0x61, 0x6e, 0x74, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x22, 0x65, 0x0a, 0x11, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x65, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x66, 0x65, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x6b, 0x77, 0x69, 0x6c, 0x2d, 0x64, 0x62, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x72, 0x70,
	0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31,
	0x3b, 0x74, 0x78, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kwil_tx_v1_tx_proto_rawDescOnce sync.Once
	file_kwil_tx_v1_tx_proto_rawDescData = file_kwil_tx_v1_tx_proto_rawDesc
)

func file_kwil_tx_v1_tx_proto_rawDescGZIP() []byte {
	file_kwil_tx_v1_tx_proto_rawDescOnce.Do(func() {
		file_kwil_tx_v1_tx_proto_rawDescData = protoimpl.X.CompressGZIP(file_kwil_tx_v1_tx_proto_rawDescData)
	})
	return file_kwil_tx_v1_tx_proto_rawDescData
}

var file_kwil_tx_v1_tx_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_kwil_tx_v1_tx_proto_goTypes = []interface{}{
	(*Transaction)(nil),       // 0: tx.Transaction
	(*Signature)(nil),         // 1: tx.Signature
	(*TransactionResult)(nil), // 2: tx.TransactionResult
	(*TransactionStatus)(nil), // 3: tx.TransactionStatus
	(*Transaction_Body)(nil),  // 4: tx.Transaction.Body
}
var file_kwil_tx_v1_tx_proto_depIdxs = []int32{
	4, // 0: tx.Transaction.body:type_name -> tx.Transaction.Body
	1, // 1: tx.Transaction.signature:type_name -> tx.Signature
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_kwil_tx_v1_tx_proto_init() }
func file_kwil_tx_v1_tx_proto_init() {
	if File_kwil_tx_v1_tx_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kwil_tx_v1_tx_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kwil_tx_v1_tx_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Signature); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kwil_tx_v1_tx_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kwil_tx_v1_tx_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kwil_tx_v1_tx_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction_Body); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kwil_tx_v1_tx_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kwil_tx_v1_tx_proto_goTypes,
		DependencyIndexes: file_kwil_tx_v1_tx_proto_depIdxs,
		MessageInfos:      file_kwil_tx_v1_tx_proto_msgTypes,
	}.Build()
	File_kwil_tx_v1_tx_proto = out.File
	file_kwil_tx_v1_tx_proto_rawDesc = nil
	file_kwil_tx_v1_tx_proto_goTypes = nil
	file_kwil_tx_v1_tx_proto_depIdxs = nil
}
