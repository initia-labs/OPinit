// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: opinit/opchild/v1/tx.proto

package opchildv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Msg_ExecuteMessages_FullMethodName         = "/opinit.opchild.v1.Msg/ExecuteMessages"
	Msg_SetBridgeInfo_FullMethodName           = "/opinit.opchild.v1.Msg/SetBridgeInfo"
	Msg_FinalizeTokenDeposit_FullMethodName    = "/opinit.opchild.v1.Msg/FinalizeTokenDeposit"
	Msg_InitiateTokenWithdrawal_FullMethodName = "/opinit.opchild.v1.Msg/InitiateTokenWithdrawal"
	Msg_AddValidator_FullMethodName            = "/opinit.opchild.v1.Msg/AddValidator"
	Msg_RemoveValidator_FullMethodName         = "/opinit.opchild.v1.Msg/RemoveValidator"
	Msg_UpdateParams_FullMethodName            = "/opinit.opchild.v1.Msg/UpdateParams"
	Msg_SpendFeePool_FullMethodName            = "/opinit.opchild.v1.Msg/SpendFeePool"
	Msg_UpdateOracle_FullMethodName            = "/opinit.opchild.v1.Msg/UpdateOracle"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Msg defines the rollup Msg service.
type MsgClient interface {
	// ExecuteMessages defines a rpc handler method for MsgExecuteMessages.
	ExecuteMessages(ctx context.Context, in *MsgExecuteMessages, opts ...grpc.CallOption) (*MsgExecuteMessagesResponse, error)
	// SetBridgeInfo defines a rpc handler method for MsgSetBridgeInfo.
	SetBridgeInfo(ctx context.Context, in *MsgSetBridgeInfo, opts ...grpc.CallOption) (*MsgSetBridgeInfoResponse, error)
	// FinalizeTokenDeposit defines a rpc handler method for MsgFinalizeTokenDeposit.
	FinalizeTokenDeposit(ctx context.Context, in *MsgFinalizeTokenDeposit, opts ...grpc.CallOption) (*MsgFinalizeTokenDepositResponse, error)
	// InitiateTokenWithdrawal defines a user facing l2 => l1 token transfer interface.
	InitiateTokenWithdrawal(ctx context.Context, in *MsgInitiateTokenWithdrawal, opts ...grpc.CallOption) (*MsgInitiateTokenWithdrawalResponse, error)
	// AddValidator defines a rpc handler method for MsgAddValidator.
	AddValidator(ctx context.Context, in *MsgAddValidator, opts ...grpc.CallOption) (*MsgAddValidatorResponse, error)
	// RemoveValidator defines a rpc handler method for MsgRemoveValidator.
	RemoveValidator(ctx context.Context, in *MsgRemoveValidator, opts ...grpc.CallOption) (*MsgRemoveValidatorResponse, error)
	// UpdateParams defines an operation for updating the
	// x/opchild module parameters.
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
	// SpendFeePool defines an operation that spend fee pool to a recipient.
	SpendFeePool(ctx context.Context, in *MsgSpendFeePool, opts ...grpc.CallOption) (*MsgSpendFeePoolResponse, error)
	// UpdateOracle defines an operation that update oracle prices.
	UpdateOracle(ctx context.Context, in *MsgUpdateOracle, opts ...grpc.CallOption) (*MsgUpdateOracleResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) ExecuteMessages(ctx context.Context, in *MsgExecuteMessages, opts ...grpc.CallOption) (*MsgExecuteMessagesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgExecuteMessagesResponse)
	err := c.cc.Invoke(ctx, Msg_ExecuteMessages_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SetBridgeInfo(ctx context.Context, in *MsgSetBridgeInfo, opts ...grpc.CallOption) (*MsgSetBridgeInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgSetBridgeInfoResponse)
	err := c.cc.Invoke(ctx, Msg_SetBridgeInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) FinalizeTokenDeposit(ctx context.Context, in *MsgFinalizeTokenDeposit, opts ...grpc.CallOption) (*MsgFinalizeTokenDepositResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgFinalizeTokenDepositResponse)
	err := c.cc.Invoke(ctx, Msg_FinalizeTokenDeposit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) InitiateTokenWithdrawal(ctx context.Context, in *MsgInitiateTokenWithdrawal, opts ...grpc.CallOption) (*MsgInitiateTokenWithdrawalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgInitiateTokenWithdrawalResponse)
	err := c.cc.Invoke(ctx, Msg_InitiateTokenWithdrawal_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) AddValidator(ctx context.Context, in *MsgAddValidator, opts ...grpc.CallOption) (*MsgAddValidatorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgAddValidatorResponse)
	err := c.cc.Invoke(ctx, Msg_AddValidator_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RemoveValidator(ctx context.Context, in *MsgRemoveValidator, opts ...grpc.CallOption) (*MsgRemoveValidatorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgRemoveValidatorResponse)
	err := c.cc.Invoke(ctx, Msg_RemoveValidator_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateParams_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) SpendFeePool(ctx context.Context, in *MsgSpendFeePool, opts ...grpc.CallOption) (*MsgSpendFeePoolResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgSpendFeePoolResponse)
	err := c.cc.Invoke(ctx, Msg_SpendFeePool_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateOracle(ctx context.Context, in *MsgUpdateOracle, opts ...grpc.CallOption) (*MsgUpdateOracleResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MsgUpdateOracleResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateOracle_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
//
// Msg defines the rollup Msg service.
type MsgServer interface {
	// ExecuteMessages defines a rpc handler method for MsgExecuteMessages.
	ExecuteMessages(context.Context, *MsgExecuteMessages) (*MsgExecuteMessagesResponse, error)
	// SetBridgeInfo defines a rpc handler method for MsgSetBridgeInfo.
	SetBridgeInfo(context.Context, *MsgSetBridgeInfo) (*MsgSetBridgeInfoResponse, error)
	// FinalizeTokenDeposit defines a rpc handler method for MsgFinalizeTokenDeposit.
	FinalizeTokenDeposit(context.Context, *MsgFinalizeTokenDeposit) (*MsgFinalizeTokenDepositResponse, error)
	// InitiateTokenWithdrawal defines a user facing l2 => l1 token transfer interface.
	InitiateTokenWithdrawal(context.Context, *MsgInitiateTokenWithdrawal) (*MsgInitiateTokenWithdrawalResponse, error)
	// AddValidator defines a rpc handler method for MsgAddValidator.
	AddValidator(context.Context, *MsgAddValidator) (*MsgAddValidatorResponse, error)
	// RemoveValidator defines a rpc handler method for MsgRemoveValidator.
	RemoveValidator(context.Context, *MsgRemoveValidator) (*MsgRemoveValidatorResponse, error)
	// UpdateParams defines an operation for updating the
	// x/opchild module parameters.
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	// SpendFeePool defines an operation that spend fee pool to a recipient.
	SpendFeePool(context.Context, *MsgSpendFeePool) (*MsgSpendFeePoolResponse, error)
	// UpdateOracle defines an operation that update oracle prices.
	UpdateOracle(context.Context, *MsgUpdateOracle) (*MsgUpdateOracleResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) ExecuteMessages(context.Context, *MsgExecuteMessages) (*MsgExecuteMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteMessages not implemented")
}
func (UnimplementedMsgServer) SetBridgeInfo(context.Context, *MsgSetBridgeInfo) (*MsgSetBridgeInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetBridgeInfo not implemented")
}
func (UnimplementedMsgServer) FinalizeTokenDeposit(context.Context, *MsgFinalizeTokenDeposit) (*MsgFinalizeTokenDepositResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinalizeTokenDeposit not implemented")
}
func (UnimplementedMsgServer) InitiateTokenWithdrawal(context.Context, *MsgInitiateTokenWithdrawal) (*MsgInitiateTokenWithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitiateTokenWithdrawal not implemented")
}
func (UnimplementedMsgServer) AddValidator(context.Context, *MsgAddValidator) (*MsgAddValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddValidator not implemented")
}
func (UnimplementedMsgServer) RemoveValidator(context.Context, *MsgRemoveValidator) (*MsgRemoveValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveValidator not implemented")
}
func (UnimplementedMsgServer) UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}
func (UnimplementedMsgServer) SpendFeePool(context.Context, *MsgSpendFeePool) (*MsgSpendFeePoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SpendFeePool not implemented")
}
func (UnimplementedMsgServer) UpdateOracle(context.Context, *MsgUpdateOracle) (*MsgUpdateOracleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOracle not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_ExecuteMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgExecuteMessages)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ExecuteMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ExecuteMessages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ExecuteMessages(ctx, req.(*MsgExecuteMessages))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SetBridgeInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetBridgeInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SetBridgeInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SetBridgeInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SetBridgeInfo(ctx, req.(*MsgSetBridgeInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_FinalizeTokenDeposit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFinalizeTokenDeposit)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FinalizeTokenDeposit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_FinalizeTokenDeposit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FinalizeTokenDeposit(ctx, req.(*MsgFinalizeTokenDeposit))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_InitiateTokenWithdrawal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgInitiateTokenWithdrawal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).InitiateTokenWithdrawal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_InitiateTokenWithdrawal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).InitiateTokenWithdrawal(ctx, req.(*MsgInitiateTokenWithdrawal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_AddValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddValidator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddValidator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_AddValidator_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddValidator(ctx, req.(*MsgAddValidator))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveValidator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RemoveValidator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RemoveValidator_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RemoveValidator(ctx, req.(*MsgRemoveValidator))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateParams_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SpendFeePool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSpendFeePool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SpendFeePool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_SpendFeePool_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SpendFeePool(ctx, req.(*MsgSpendFeePool))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateOracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateOracle)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateOracle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateOracle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateOracle(ctx, req.(*MsgUpdateOracle))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "opinit.opchild.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecuteMessages",
			Handler:    _Msg_ExecuteMessages_Handler,
		},
		{
			MethodName: "SetBridgeInfo",
			Handler:    _Msg_SetBridgeInfo_Handler,
		},
		{
			MethodName: "FinalizeTokenDeposit",
			Handler:    _Msg_FinalizeTokenDeposit_Handler,
		},
		{
			MethodName: "InitiateTokenWithdrawal",
			Handler:    _Msg_InitiateTokenWithdrawal_Handler,
		},
		{
			MethodName: "AddValidator",
			Handler:    _Msg_AddValidator_Handler,
		},
		{
			MethodName: "RemoveValidator",
			Handler:    _Msg_RemoveValidator_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
		{
			MethodName: "SpendFeePool",
			Handler:    _Msg_SpendFeePool_Handler,
		},
		{
			MethodName: "UpdateOracle",
			Handler:    _Msg_UpdateOracle_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "opinit/opchild/v1/tx.proto",
}
