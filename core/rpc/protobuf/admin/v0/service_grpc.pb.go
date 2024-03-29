// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: kwil/admin/v0/service.proto

package admpb

import (
	context "context"
	v1 "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	AdminService_Version_FullMethodName          = "/admin.AdminService/Version"
	AdminService_Status_FullMethodName           = "/admin.AdminService/Status"
	AdminService_Peers_FullMethodName            = "/admin.AdminService/Peers"
	AdminService_Approve_FullMethodName          = "/admin.AdminService/Approve"
	AdminService_Join_FullMethodName             = "/admin.AdminService/Join"
	AdminService_Leave_FullMethodName            = "/admin.AdminService/Leave"
	AdminService_Remove_FullMethodName           = "/admin.AdminService/Remove"
	AdminService_JoinStatus_FullMethodName       = "/admin.AdminService/JoinStatus"
	AdminService_ListValidators_FullMethodName   = "/admin.AdminService/ListValidators"
	AdminService_ListPendingJoins_FullMethodName = "/admin.AdminService/ListPendingJoins"
	AdminService_GetConfig_FullMethodName        = "/admin.AdminService/GetConfig"
)

// AdminServiceClient is the client API for AdminService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminServiceClient interface {
	Version(ctx context.Context, in *VersionRequest, opts ...grpc.CallOption) (*VersionResponse, error)
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	Peers(ctx context.Context, in *PeersRequest, opts ...grpc.CallOption) (*PeersResponse, error)
	Approve(ctx context.Context, in *ApproveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error)
	Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error)
	Leave(ctx context.Context, in *LeaveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error)
	Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error)
	JoinStatus(ctx context.Context, in *JoinStatusRequest, opts ...grpc.CallOption) (*JoinStatusResponse, error)
	ListValidators(ctx context.Context, in *ListValidatorsRequest, opts ...grpc.CallOption) (*ListValidatorsResponse, error)
	ListPendingJoins(ctx context.Context, in *ListJoinRequestsRequest, opts ...grpc.CallOption) (*ListJoinRequestsResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
}

type adminServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminServiceClient(cc grpc.ClientConnInterface) AdminServiceClient {
	return &adminServiceClient{cc}
}

func (c *adminServiceClient) Version(ctx context.Context, in *VersionRequest, opts ...grpc.CallOption) (*VersionResponse, error) {
	out := new(VersionResponse)
	err := c.cc.Invoke(ctx, AdminService_Version_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, AdminService_Status_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Peers(ctx context.Context, in *PeersRequest, opts ...grpc.CallOption) (*PeersResponse, error) {
	out := new(PeersResponse)
	err := c.cc.Invoke(ctx, AdminService_Peers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Approve(ctx context.Context, in *ApproveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error) {
	out := new(v1.BroadcastResponse)
	err := c.cc.Invoke(ctx, AdminService_Approve_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Join(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error) {
	out := new(v1.BroadcastResponse)
	err := c.cc.Invoke(ctx, AdminService_Join_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Leave(ctx context.Context, in *LeaveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error) {
	out := new(v1.BroadcastResponse)
	err := c.cc.Invoke(ctx, AdminService_Leave_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*v1.BroadcastResponse, error) {
	out := new(v1.BroadcastResponse)
	err := c.cc.Invoke(ctx, AdminService_Remove_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) JoinStatus(ctx context.Context, in *JoinStatusRequest, opts ...grpc.CallOption) (*JoinStatusResponse, error) {
	out := new(JoinStatusResponse)
	err := c.cc.Invoke(ctx, AdminService_JoinStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) ListValidators(ctx context.Context, in *ListValidatorsRequest, opts ...grpc.CallOption) (*ListValidatorsResponse, error) {
	out := new(ListValidatorsResponse)
	err := c.cc.Invoke(ctx, AdminService_ListValidators_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) ListPendingJoins(ctx context.Context, in *ListJoinRequestsRequest, opts ...grpc.CallOption) (*ListJoinRequestsResponse, error) {
	out := new(ListJoinRequestsResponse)
	err := c.cc.Invoke(ctx, AdminService_ListPendingJoins_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, AdminService_GetConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminServiceServer is the server API for AdminService service.
// All implementations must embed UnimplementedAdminServiceServer
// for forward compatibility
type AdminServiceServer interface {
	Version(context.Context, *VersionRequest) (*VersionResponse, error)
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	Peers(context.Context, *PeersRequest) (*PeersResponse, error)
	Approve(context.Context, *ApproveRequest) (*v1.BroadcastResponse, error)
	Join(context.Context, *JoinRequest) (*v1.BroadcastResponse, error)
	Leave(context.Context, *LeaveRequest) (*v1.BroadcastResponse, error)
	Remove(context.Context, *RemoveRequest) (*v1.BroadcastResponse, error)
	JoinStatus(context.Context, *JoinStatusRequest) (*JoinStatusResponse, error)
	ListValidators(context.Context, *ListValidatorsRequest) (*ListValidatorsResponse, error)
	ListPendingJoins(context.Context, *ListJoinRequestsRequest) (*ListJoinRequestsResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	mustEmbedUnimplementedAdminServiceServer()
}

// UnimplementedAdminServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAdminServiceServer struct {
}

func (UnimplementedAdminServiceServer) Version(context.Context, *VersionRequest) (*VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedAdminServiceServer) Status(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedAdminServiceServer) Peers(context.Context, *PeersRequest) (*PeersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Peers not implemented")
}
func (UnimplementedAdminServiceServer) Approve(context.Context, *ApproveRequest) (*v1.BroadcastResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Approve not implemented")
}
func (UnimplementedAdminServiceServer) Join(context.Context, *JoinRequest) (*v1.BroadcastResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedAdminServiceServer) Leave(context.Context, *LeaveRequest) (*v1.BroadcastResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Leave not implemented")
}
func (UnimplementedAdminServiceServer) Remove(context.Context, *RemoveRequest) (*v1.BroadcastResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remove not implemented")
}
func (UnimplementedAdminServiceServer) JoinStatus(context.Context, *JoinStatusRequest) (*JoinStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinStatus not implemented")
}
func (UnimplementedAdminServiceServer) ListValidators(context.Context, *ListValidatorsRequest) (*ListValidatorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListValidators not implemented")
}
func (UnimplementedAdminServiceServer) ListPendingJoins(context.Context, *ListJoinRequestsRequest) (*ListJoinRequestsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPendingJoins not implemented")
}
func (UnimplementedAdminServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedAdminServiceServer) mustEmbedUnimplementedAdminServiceServer() {}

// UnsafeAdminServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminServiceServer will
// result in compilation errors.
type UnsafeAdminServiceServer interface {
	mustEmbedUnimplementedAdminServiceServer()
}

func RegisterAdminServiceServer(s grpc.ServiceRegistrar, srv AdminServiceServer) {
	s.RegisterService(&AdminService_ServiceDesc, srv)
}

func _AdminService_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Version_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Version(ctx, req.(*VersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Status_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Peers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Peers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Peers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Peers(ctx, req.(*PeersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Approve_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApproveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Approve(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Approve_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Approve(ctx, req.(*ApproveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Join(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Join_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Join(ctx, req.(*JoinRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Leave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Leave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Leave_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Leave(ctx, req.(*LeaveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_Remove_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).Remove(ctx, req.(*RemoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_JoinStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).JoinStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_JoinStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).JoinStatus(ctx, req.(*JoinStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_ListValidators_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListValidatorsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).ListValidators(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_ListValidators_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).ListValidators(ctx, req.(*ListValidatorsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_ListPendingJoins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListJoinRequestsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).ListPendingJoins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_ListPendingJoins_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).ListPendingJoins(ctx, req.(*ListJoinRequestsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AdminService_GetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AdminService_ServiceDesc is the grpc.ServiceDesc for AdminService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AdminService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "admin.AdminService",
	HandlerType: (*AdminServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _AdminService_Version_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _AdminService_Status_Handler,
		},
		{
			MethodName: "Peers",
			Handler:    _AdminService_Peers_Handler,
		},
		{
			MethodName: "Approve",
			Handler:    _AdminService_Approve_Handler,
		},
		{
			MethodName: "Join",
			Handler:    _AdminService_Join_Handler,
		},
		{
			MethodName: "Leave",
			Handler:    _AdminService_Leave_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _AdminService_Remove_Handler,
		},
		{
			MethodName: "JoinStatus",
			Handler:    _AdminService_JoinStatus_Handler,
		},
		{
			MethodName: "ListValidators",
			Handler:    _AdminService_ListValidators_Handler,
		},
		{
			MethodName: "ListPendingJoins",
			Handler:    _AdminService_ListPendingJoins_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _AdminService_GetConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kwil/admin/v0/service.proto",
}
