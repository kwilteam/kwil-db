// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: kwil/tx/v1/account.proto

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

type AccountStatus int32

const (
	AccountStatus_latest  AccountStatus = 0
	AccountStatus_pending AccountStatus = 1
)

// Enum value maps for AccountStatus.
var (
	AccountStatus_name = map[int32]string{
		0: "latest",
		1: "pending",
	}
	AccountStatus_value = map[string]int32{
		"latest":  0,
		"pending": 1,
	}
)

func (x AccountStatus) Enum() *AccountStatus {
	p := new(AccountStatus)
	*p = x
	return p
}

func (x AccountStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AccountStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_kwil_tx_v1_account_proto_enumTypes[0].Descriptor()
}

func (AccountStatus) Type() protoreflect.EnumType {
	return &file_kwil_tx_v1_account_proto_enumTypes[0]
}

func (x AccountStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AccountStatus.Descriptor instead.
func (AccountStatus) EnumDescriptor() ([]byte, []int) {
	return file_kwil_tx_v1_account_proto_rawDescGZIP(), []int{0}
}

type Account struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey []byte `protobuf:"bytes,1,opt,name=public_key,proto3" json:"public_key,omitempty"`
	Balance   string `protobuf:"bytes,2,opt,name=balance,proto3" json:"balance,omitempty"`
	Nonce     int64  `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
}

func (x *Account) Reset() {
	*x = Account{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_account_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Account) ProtoMessage() {}

func (x *Account) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_account_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Account.ProtoReflect.Descriptor instead.
func (*Account) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_account_proto_rawDescGZIP(), []int{0}
}

func (x *Account) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *Account) GetBalance() string {
	if x != nil {
		return x.Balance
	}
	return ""
}

func (x *Account) GetNonce() int64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

type GetAccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey []byte         `protobuf:"bytes,1,opt,name=public_key,proto3" json:"public_key,omitempty"`
	Status    *AccountStatus `protobuf:"varint,2,opt,name=status,proto3,enum=tx.AccountStatus,oneof" json:"status,omitempty"`
}

func (x *GetAccountRequest) Reset() {
	*x = GetAccountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_account_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAccountRequest) ProtoMessage() {}

func (x *GetAccountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_account_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAccountRequest.ProtoReflect.Descriptor instead.
func (*GetAccountRequest) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_account_proto_rawDescGZIP(), []int{1}
}

func (x *GetAccountRequest) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *GetAccountRequest) GetStatus() AccountStatus {
	if x != nil && x.Status != nil {
		return *x.Status
	}
	return AccountStatus_latest
}

type GetAccountResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account *Account `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
}

func (x *GetAccountResponse) Reset() {
	*x = GetAccountResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kwil_tx_v1_account_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAccountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAccountResponse) ProtoMessage() {}

func (x *GetAccountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kwil_tx_v1_account_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAccountResponse.ProtoReflect.Descriptor instead.
func (*GetAccountResponse) Descriptor() ([]byte, []int) {
	return file_kwil_tx_v1_account_proto_rawDescGZIP(), []int{2}
}

func (x *GetAccountResponse) GetAccount() *Account {
	if x != nil {
		return x.Account
	}
	return nil
}

var File_kwil_tx_v1_account_proto protoreflect.FileDescriptor

var file_kwil_tx_v1_account_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6b, 0x77, 0x69, 0x6c, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x74, 0x78, 0x22, 0x59,
	0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x62, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x22, 0x6e, 0x0a, 0x11, 0x47, 0x65, 0x74,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e,
	0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x12, 0x2e,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11,
	0x2e, 0x74, 0x78, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x48, 0x00, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x88, 0x01, 0x01, 0x42, 0x09,
	0x0a, 0x07, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x3b, 0x0a, 0x12, 0x47, 0x65, 0x74,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x25, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0b, 0x2e, 0x74, 0x78, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x07, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2a, 0x28, 0x0a, 0x0d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0a, 0x0a, 0x06, 0x6c, 0x61, 0x74, 0x65, 0x73,
	0x74, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x10, 0x01,
	0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b,
	0x77, 0x69, 0x6c, 0x74, 0x65, 0x61, 0x6d, 0x2f, 0x6b, 0x77, 0x69, 0x6c, 0x2d, 0x64, 0x62, 0x2f,
	0x63, 0x6f, 0x72, 0x65, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x78, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x78, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kwil_tx_v1_account_proto_rawDescOnce sync.Once
	file_kwil_tx_v1_account_proto_rawDescData = file_kwil_tx_v1_account_proto_rawDesc
)

func file_kwil_tx_v1_account_proto_rawDescGZIP() []byte {
	file_kwil_tx_v1_account_proto_rawDescOnce.Do(func() {
		file_kwil_tx_v1_account_proto_rawDescData = protoimpl.X.CompressGZIP(file_kwil_tx_v1_account_proto_rawDescData)
	})
	return file_kwil_tx_v1_account_proto_rawDescData
}

var file_kwil_tx_v1_account_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_kwil_tx_v1_account_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_kwil_tx_v1_account_proto_goTypes = []interface{}{
	(AccountStatus)(0),         // 0: tx.AccountStatus
	(*Account)(nil),            // 1: tx.Account
	(*GetAccountRequest)(nil),  // 2: tx.GetAccountRequest
	(*GetAccountResponse)(nil), // 3: tx.GetAccountResponse
}
var file_kwil_tx_v1_account_proto_depIdxs = []int32{
	0, // 0: tx.GetAccountRequest.status:type_name -> tx.AccountStatus
	1, // 1: tx.GetAccountResponse.account:type_name -> tx.Account
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_kwil_tx_v1_account_proto_init() }
func file_kwil_tx_v1_account_proto_init() {
	if File_kwil_tx_v1_account_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kwil_tx_v1_account_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Account); i {
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
		file_kwil_tx_v1_account_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAccountRequest); i {
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
		file_kwil_tx_v1_account_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAccountResponse); i {
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
	file_kwil_tx_v1_account_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kwil_tx_v1_account_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kwil_tx_v1_account_proto_goTypes,
		DependencyIndexes: file_kwil_tx_v1_account_proto_depIdxs,
		EnumInfos:         file_kwil_tx_v1_account_proto_enumTypes,
		MessageInfos:      file_kwil_tx_v1_account_proto_msgTypes,
	}.Build()
	File_kwil_tx_v1_account_proto = out.File
	file_kwil_tx_v1_account_proto_rawDesc = nil
	file_kwil_tx_v1_account_proto_goTypes = nil
	file_kwil_tx_v1_account_proto_depIdxs = nil
}
