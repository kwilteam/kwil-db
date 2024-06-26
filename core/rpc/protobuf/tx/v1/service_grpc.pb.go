// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: kwil/tx/v1/service.proto

package txpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	TxService_ChainInfo_FullMethodName     = "/tx.TxService/ChainInfo"
	TxService_Broadcast_FullMethodName     = "/tx.TxService/Broadcast"
	TxService_EstimatePrice_FullMethodName = "/tx.TxService/EstimatePrice"
	TxService_Query_FullMethodName         = "/tx.TxService/Query"
	TxService_GetAccount_FullMethodName    = "/tx.TxService/GetAccount"
	TxService_Ping_FullMethodName          = "/tx.TxService/Ping"
	TxService_ListDatabases_FullMethodName = "/tx.TxService/ListDatabases"
	TxService_GetSchema_FullMethodName     = "/tx.TxService/GetSchema"
	TxService_Call_FullMethodName          = "/tx.TxService/Call"
	TxService_TxQuery_FullMethodName       = "/tx.TxService/TxQuery"
)

// TxServiceClient is the client API for TxService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TxServiceClient interface {
	ChainInfo(ctx context.Context, in *ChainInfoRequest, opts ...grpc.CallOption) (*ChainInfoResponse, error)
	Broadcast(ctx context.Context, in *BroadcastRequest, opts ...grpc.CallOption) (*BroadcastResponse, error)
	EstimatePrice(ctx context.Context, in *EstimatePriceRequest, opts ...grpc.CallOption) (*EstimatePriceResponse, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error)
	GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*GetAccountResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	ListDatabases(ctx context.Context, in *ListDatabasesRequest, opts ...grpc.CallOption) (*ListDatabasesResponse, error)
	GetSchema(ctx context.Context, in *GetSchemaRequest, opts ...grpc.CallOption) (*GetSchemaResponse, error)
	Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error)
	TxQuery(ctx context.Context, in *TxQueryRequest, opts ...grpc.CallOption) (*TxQueryResponse, error)
}

type txServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTxServiceClient(cc grpc.ClientConnInterface) TxServiceClient {
	return &txServiceClient{cc}
}

func (c *txServiceClient) ChainInfo(ctx context.Context, in *ChainInfoRequest, opts ...grpc.CallOption) (*ChainInfoResponse, error) {
	out := new(ChainInfoResponse)
	err := c.cc.Invoke(ctx, TxService_ChainInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) Broadcast(ctx context.Context, in *BroadcastRequest, opts ...grpc.CallOption) (*BroadcastResponse, error) {
	out := new(BroadcastResponse)
	err := c.cc.Invoke(ctx, TxService_Broadcast_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) EstimatePrice(ctx context.Context, in *EstimatePriceRequest, opts ...grpc.CallOption) (*EstimatePriceResponse, error) {
	out := new(EstimatePriceResponse)
	err := c.cc.Invoke(ctx, TxService_EstimatePrice_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := c.cc.Invoke(ctx, TxService_Query_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*GetAccountResponse, error) {
	out := new(GetAccountResponse)
	err := c.cc.Invoke(ctx, TxService_GetAccount_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, TxService_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) ListDatabases(ctx context.Context, in *ListDatabasesRequest, opts ...grpc.CallOption) (*ListDatabasesResponse, error) {
	out := new(ListDatabasesResponse)
	err := c.cc.Invoke(ctx, TxService_ListDatabases_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) GetSchema(ctx context.Context, in *GetSchemaRequest, opts ...grpc.CallOption) (*GetSchemaResponse, error) {
	out := new(GetSchemaResponse)
	err := c.cc.Invoke(ctx, TxService_GetSchema_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error) {
	out := new(CallResponse)
	err := c.cc.Invoke(ctx, TxService_Call_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *txServiceClient) TxQuery(ctx context.Context, in *TxQueryRequest, opts ...grpc.CallOption) (*TxQueryResponse, error) {
	out := new(TxQueryResponse)
	err := c.cc.Invoke(ctx, TxService_TxQuery_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TxServiceServer is the server API for TxService service.
// All implementations must embed UnimplementedTxServiceServer
// for forward compatibility
type TxServiceServer interface {
	ChainInfo(context.Context, *ChainInfoRequest) (*ChainInfoResponse, error)
	Broadcast(context.Context, *BroadcastRequest) (*BroadcastResponse, error)
	EstimatePrice(context.Context, *EstimatePriceRequest) (*EstimatePriceResponse, error)
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
	GetAccount(context.Context, *GetAccountRequest) (*GetAccountResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	ListDatabases(context.Context, *ListDatabasesRequest) (*ListDatabasesResponse, error)
	GetSchema(context.Context, *GetSchemaRequest) (*GetSchemaResponse, error)
	Call(context.Context, *CallRequest) (*CallResponse, error)
	TxQuery(context.Context, *TxQueryRequest) (*TxQueryResponse, error)
	mustEmbedUnimplementedTxServiceServer()
}

// UnimplementedTxServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTxServiceServer struct {
}

func (UnimplementedTxServiceServer) ChainInfo(context.Context, *ChainInfoRequest) (*ChainInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChainInfo not implemented")
}
func (UnimplementedTxServiceServer) Broadcast(context.Context, *BroadcastRequest) (*BroadcastResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Broadcast not implemented")
}
func (UnimplementedTxServiceServer) EstimatePrice(context.Context, *EstimatePriceRequest) (*EstimatePriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EstimatePrice not implemented")
}
func (UnimplementedTxServiceServer) Query(context.Context, *QueryRequest) (*QueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}
func (UnimplementedTxServiceServer) GetAccount(context.Context, *GetAccountRequest) (*GetAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}
func (UnimplementedTxServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedTxServiceServer) ListDatabases(context.Context, *ListDatabasesRequest) (*ListDatabasesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDatabases not implemented")
}
func (UnimplementedTxServiceServer) GetSchema(context.Context, *GetSchemaRequest) (*GetSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSchema not implemented")
}
func (UnimplementedTxServiceServer) Call(context.Context, *CallRequest) (*CallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedTxServiceServer) TxQuery(context.Context, *TxQueryRequest) (*TxQueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TxQuery not implemented")
}
func (UnimplementedTxServiceServer) mustEmbedUnimplementedTxServiceServer() {}

// UnsafeTxServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TxServiceServer will
// result in compilation errors.
type UnsafeTxServiceServer interface {
	mustEmbedUnimplementedTxServiceServer()
}

func RegisterTxServiceServer(s grpc.ServiceRegistrar, srv TxServiceServer) {
	s.RegisterService(&TxService_ServiceDesc, srv)
}

func _TxService_ChainInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChainInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).ChainInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_ChainInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).ChainInfo(ctx, req.(*ChainInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_Broadcast_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadcastRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).Broadcast(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_Broadcast_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).Broadcast(ctx, req.(*BroadcastRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_EstimatePrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EstimatePriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).EstimatePrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_EstimatePrice_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).EstimatePrice(ctx, req.(*EstimatePriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_Query_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_GetAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).GetAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_GetAccount_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).GetAccount(ctx, req.(*GetAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_ListDatabases_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDatabasesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).ListDatabases(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_ListDatabases_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).ListDatabases(ctx, req.(*ListDatabasesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_GetSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).GetSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_GetSchema_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).GetSchema(ctx, req.(*GetSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_Call_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).Call(ctx, req.(*CallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TxService_TxQuery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxQueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TxServiceServer).TxQuery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TxService_TxQuery_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TxServiceServer).TxQuery(ctx, req.(*TxQueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TxService_ServiceDesc is the grpc.ServiceDesc for TxService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TxService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tx.TxService",
	HandlerType: (*TxServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ChainInfo",
			Handler:    _TxService_ChainInfo_Handler,
		},
		{
			MethodName: "Broadcast",
			Handler:    _TxService_Broadcast_Handler,
		},
		{
			MethodName: "EstimatePrice",
			Handler:    _TxService_EstimatePrice_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _TxService_Query_Handler,
		},
		{
			MethodName: "GetAccount",
			Handler:    _TxService_GetAccount_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _TxService_Ping_Handler,
		},
		{
			MethodName: "ListDatabases",
			Handler:    _TxService_ListDatabases_Handler,
		},
		{
			MethodName: "GetSchema",
			Handler:    _TxService_GetSchema_Handler,
		},
		{
			MethodName: "Call",
			Handler:    _TxService_Call_Handler,
		},
		{
			MethodName: "TxQuery",
			Handler:    _TxService_TxQuery_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kwil/tx/v1/service.proto",
}
