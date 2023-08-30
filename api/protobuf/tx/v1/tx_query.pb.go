// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.4
// source: kwil/tx/v1/tx_query.proto

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

type TxQueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash []byte `protobuf:"bytes,1,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
}

func (x *TxQueryRequest) Reset() {
	*x = TxQueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxQueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxQueryRequest) ProtoMessage() {}

func (x *TxQueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxQueryRequest.ProtoReflect.Descriptor instead.
func (*TxQueryRequest) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_query_proto_rawDescGZIP(), []int{0}
}

func (x *TxQueryRequest) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

type TxQueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash     []byte             `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Height   uint64             `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Tx       *Transaction       `protobuf:"bytes,3,opt,name=tx,proto3" json:"tx,omitempty"`
	TxResult *TransactionResult `protobuf:"bytes,4,opt,name=tx_result,json=txResult,proto3" json:"tx_result,omitempty"`
}

func (x *TxQueryResponse) Reset() {
	*x = TxQueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_tx_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxQueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxQueryResponse) ProtoMessage() {}

func (x *TxQueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_tx_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxQueryResponse.ProtoReflect.Descriptor instead.
func (*TxQueryResponse) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_tx_query_proto_rawDescGZIP(), []int{1}
}

func (x *TxQueryResponse) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *TxQueryResponse) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *TxQueryResponse) GetTx() *Transaction {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *TxQueryResponse) GetTxResult() *TransactionResult {
	if x != nil {
		return x.TxResult
	}
	return nil
}

var File_kwil_tx_v1_tx_query_proto protoreflect.FileDescriptor

var file_kwil_tx_v1_tx_query_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x78, 0x5f,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x74, 0x78, 0x1a,
	0x13, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x78, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x29, 0x0a, 0x0e, 0x54, 0x78, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x22,
	0x92, 0x01, 0x0a, 0x0f, 0x54, 0x78, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x1f, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x74, 0x78,
	0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x02, 0x74, 0x78,
	0x12, 0x32, 0x0a, 0x09, 0x74, 0x78, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x78, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x08, 0x74, 0x78, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c, 0x74, 0x65, 0x61, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c,
	0x2d, 0x64, 0x62, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x78, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_kwil_tx_v1_tx_query_proto_rawDescOnce sync.Once
	file_kwil_tx_v1_tx_query_proto_rawDescData = file_kwil_tx_v1_tx_query_proto_rawDesc
)

func file_kwil_tx_v1_tx_query_proto_rawDescGZIP() []byte {
	file_kwil_tx_v1_tx_query_proto_rawDescOnce.Do(func() {
		file_kwil_tx_v1_tx_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_kwil_tx_v1_tx_query_proto_rawDescData)
	})
	return file_kwil_tx_v1_tx_query_proto_rawDescData
}

var file_kwil_tx_v1_tx_query_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_kwil_tx_v1_tx_query_proto_goTypes = []interface{}{
	(*TxQueryRequest)(nil),    // 0: tx.TxQueryRequest
	(*TxQueryResponse)(nil),   // 1: tx.TxQueryResponse
	(*Transaction)(nil),       // 2: tx.Transaction
	(*TransactionResult)(nil), // 3: tx.TransactionResult
}
var file_kwil_tx_v1_tx_query_proto_depIdxs = []int32{
	2, // 0: tx.TxQueryResponse.tx:type_name -> tx.Transaction
	3, // 1: tx.TxQueryResponse.tx_result:type_name -> tx.TransactionResult
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_kwil_tx_v1_tx_query_proto_init() }
func file_kwil_tx_v1_tx_query_proto_init() {
	if File_kwil_tx_v1_tx_query_proto != nil {
		return
	}
	file_kwil_tx_v1_tx_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_kwil_tx_v1_tx_query_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxQueryRequest); i {
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
		file_kwil_tx_v1_tx_query_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxQueryResponse); i {
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
			RawDescriptor: file_kwil_tx_v1_tx_query_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kwil_tx_v1_tx_query_proto_goTypes,
		DependencyIndexes: file_kwil_tx_v1_tx_query_proto_depIdxs,
		MessageInfos:      file_kwil_tx_v1_tx_query_proto_msgTypes,
	}.Build()
	File_kwil_tx_v1_tx_query_proto = out.File
	file_kwil_tx_v1_tx_query_proto_rawDesc = nil
	file_kwil_tx_v1_tx_query_proto_goTypes = nil
	file_kwil_tx_v1_tx_query_proto_depIdxs = nil
}
