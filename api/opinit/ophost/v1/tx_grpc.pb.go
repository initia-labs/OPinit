// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: opinit/ophost/v1/tx.proto

package ophostv1

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
	Msg_RecordBatch_FullMethodName             = "/opinit.ophost.v1.Msg/RecordBatch"
	Msg_CreateBridge_FullMethodName            = "/opinit.ophost.v1.Msg/CreateBridge"
	Msg_ProposeOutput_FullMethodName           = "/opinit.ophost.v1.Msg/ProposeOutput"
	Msg_DeleteOutput_FullMethodName            = "/opinit.ophost.v1.Msg/DeleteOutput"
	Msg_InitiateTokenDeposit_FullMethodName    = "/opinit.ophost.v1.Msg/InitiateTokenDeposit"
	Msg_FinalizeTokenWithdrawal_FullMethodName = "/opinit.ophost.v1.Msg/FinalizeTokenWithdrawal"
	Msg_UpdateProposer_FullMethodName          = "/opinit.ophost.v1.Msg/UpdateProposer"
	Msg_UpdateChallenger_FullMethodName        = "/opinit.ophost.v1.Msg/UpdateChallenger"
	Msg_UpdateBatchInfo_FullMethodName         = "/opinit.ophost.v1.Msg/UpdateBatchInfo"
	Msg_UpdateMetadata_FullMethodName          = "/opinit.ophost.v1.Msg/UpdateMetadata"
	Msg_UpdateOracleConfig_FullMethodName      = "/opinit.ophost.v1.Msg/UpdateOracleConfig"
	Msg_UpdateParams_FullMethodName            = "/opinit.ophost.v1.Msg/UpdateParams"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// RecordBatch defines a rpc handler method for MsgRecordBatch.
	RecordBatch(ctx context.Context, in *MsgRecordBatch, opts ...grpc.CallOption) (*MsgRecordBatchResponse, error)
	// CreateBridge defines a rpc handler method for MsgCreateBridge.
	CreateBridge(ctx context.Context, in *MsgCreateBridge, opts ...grpc.CallOption) (*MsgCreateBridgeResponse, error)
	// ProposeOutput defines a rpc handler method for MsgProposeOutput.
	ProposeOutput(ctx context.Context, in *MsgProposeOutput, opts ...grpc.CallOption) (*MsgProposeOutputResponse, error)
	// DeleteOutput defines a rpc handler method for MsgDeleteOutput.
	DeleteOutput(ctx context.Context, in *MsgDeleteOutput, opts ...grpc.CallOption) (*MsgDeleteOutputResponse, error)
	// InitiateTokenDeposit defines a user facing l1 => l2 token transfer interface.
	InitiateTokenDeposit(ctx context.Context, in *MsgInitiateTokenDeposit, opts ...grpc.CallOption) (*MsgInitiateTokenDepositResponse, error)
	// FinalizeTokenWithdrawal defines a user facing l2 => l1 token transfer interface.
	FinalizeTokenWithdrawal(ctx context.Context, in *MsgFinalizeTokenWithdrawal, opts ...grpc.CallOption) (*MsgFinalizeTokenWithdrawalResponse, error)
	// UpdateProposer defines a rpc handler method for MsgUpdateProposer.
	UpdateProposer(ctx context.Context, in *MsgUpdateProposer, opts ...grpc.CallOption) (*MsgUpdateProposerResponse, error)
	// UpdateChallenger defines a rpc handler method for MsgUpdateChallenger.
	UpdateChallenger(ctx context.Context, in *MsgUpdateChallenger, opts ...grpc.CallOption) (*MsgUpdateChallengerResponse, error)
	// UpdateBatchInfo defines a rpc handler method for MsgUpdateBatchInfo.
	UpdateBatchInfo(ctx context.Context, in *MsgUpdateBatchInfo, opts ...grpc.CallOption) (*MsgUpdateBatchInfoResponse, error)
	// UpdateMetadata defines a rpc handler method for MsgUpdateMetadata.
	UpdateMetadata(ctx context.Context, in *MsgUpdateMetadata, opts ...grpc.CallOption) (*MsgUpdateMetadataResponse, error)
	// UpdateOracleConfig defines a rpc handler method for MsgUpdateOracleConfig.
	UpdateOracleConfig(ctx context.Context, in *MsgUpdateOracleConfig, opts ...grpc.CallOption) (*MsgUpdateOracleConfigResponse, error)
	// UpdateParams defines an operation for updating the
	// x/opchild module parameters.
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) RecordBatch(ctx context.Context, in *MsgRecordBatch, opts ...grpc.CallOption) (*MsgRecordBatchResponse, error) {
	out := new(MsgRecordBatchResponse)
	err := c.cc.Invoke(ctx, Msg_RecordBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateBridge(ctx context.Context, in *MsgCreateBridge, opts ...grpc.CallOption) (*MsgCreateBridgeResponse, error) {
	out := new(MsgCreateBridgeResponse)
	err := c.cc.Invoke(ctx, Msg_CreateBridge_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ProposeOutput(ctx context.Context, in *MsgProposeOutput, opts ...grpc.CallOption) (*MsgProposeOutputResponse, error) {
	out := new(MsgProposeOutputResponse)
	err := c.cc.Invoke(ctx, Msg_ProposeOutput_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) DeleteOutput(ctx context.Context, in *MsgDeleteOutput, opts ...grpc.CallOption) (*MsgDeleteOutputResponse, error) {
	out := new(MsgDeleteOutputResponse)
	err := c.cc.Invoke(ctx, Msg_DeleteOutput_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) InitiateTokenDeposit(ctx context.Context, in *MsgInitiateTokenDeposit, opts ...grpc.CallOption) (*MsgInitiateTokenDepositResponse, error) {
	out := new(MsgInitiateTokenDepositResponse)
	err := c.cc.Invoke(ctx, Msg_InitiateTokenDeposit_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) FinalizeTokenWithdrawal(ctx context.Context, in *MsgFinalizeTokenWithdrawal, opts ...grpc.CallOption) (*MsgFinalizeTokenWithdrawalResponse, error) {
	out := new(MsgFinalizeTokenWithdrawalResponse)
	err := c.cc.Invoke(ctx, Msg_FinalizeTokenWithdrawal_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateProposer(ctx context.Context, in *MsgUpdateProposer, opts ...grpc.CallOption) (*MsgUpdateProposerResponse, error) {
	out := new(MsgUpdateProposerResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateProposer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateChallenger(ctx context.Context, in *MsgUpdateChallenger, opts ...grpc.CallOption) (*MsgUpdateChallengerResponse, error) {
	out := new(MsgUpdateChallengerResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateChallenger_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateBatchInfo(ctx context.Context, in *MsgUpdateBatchInfo, opts ...grpc.CallOption) (*MsgUpdateBatchInfoResponse, error) {
	out := new(MsgUpdateBatchInfoResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateBatchInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateMetadata(ctx context.Context, in *MsgUpdateMetadata, opts ...grpc.CallOption) (*MsgUpdateMetadataResponse, error) {
	out := new(MsgUpdateMetadataResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateMetadata_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateOracleConfig(ctx context.Context, in *MsgUpdateOracleConfig, opts ...grpc.CallOption) (*MsgUpdateOracleConfigResponse, error) {
	out := new(MsgUpdateOracleConfigResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateOracleConfig_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateParams_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// RecordBatch defines a rpc handler method for MsgRecordBatch.
	RecordBatch(context.Context, *MsgRecordBatch) (*MsgRecordBatchResponse, error)
	// CreateBridge defines a rpc handler method for MsgCreateBridge.
	CreateBridge(context.Context, *MsgCreateBridge) (*MsgCreateBridgeResponse, error)
	// ProposeOutput defines a rpc handler method for MsgProposeOutput.
	ProposeOutput(context.Context, *MsgProposeOutput) (*MsgProposeOutputResponse, error)
	// DeleteOutput defines a rpc handler method for MsgDeleteOutput.
	DeleteOutput(context.Context, *MsgDeleteOutput) (*MsgDeleteOutputResponse, error)
	// InitiateTokenDeposit defines a user facing l1 => l2 token transfer interface.
	InitiateTokenDeposit(context.Context, *MsgInitiateTokenDeposit) (*MsgInitiateTokenDepositResponse, error)
	// FinalizeTokenWithdrawal defines a user facing l2 => l1 token transfer interface.
	FinalizeTokenWithdrawal(context.Context, *MsgFinalizeTokenWithdrawal) (*MsgFinalizeTokenWithdrawalResponse, error)
	// UpdateProposer defines a rpc handler method for MsgUpdateProposer.
	UpdateProposer(context.Context, *MsgUpdateProposer) (*MsgUpdateProposerResponse, error)
	// UpdateChallenger defines a rpc handler method for MsgUpdateChallenger.
	UpdateChallenger(context.Context, *MsgUpdateChallenger) (*MsgUpdateChallengerResponse, error)
	// UpdateBatchInfo defines a rpc handler method for MsgUpdateBatchInfo.
	UpdateBatchInfo(context.Context, *MsgUpdateBatchInfo) (*MsgUpdateBatchInfoResponse, error)
	// UpdateMetadata defines a rpc handler method for MsgUpdateMetadata.
	UpdateMetadata(context.Context, *MsgUpdateMetadata) (*MsgUpdateMetadataResponse, error)
	// UpdateOracleConfig defines a rpc handler method for MsgUpdateOracleConfig.
	UpdateOracleConfig(context.Context, *MsgUpdateOracleConfig) (*MsgUpdateOracleConfigResponse, error)
	// UpdateParams defines an operation for updating the
	// x/opchild module parameters.
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) RecordBatch(context.Context, *MsgRecordBatch) (*MsgRecordBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecordBatch not implemented")
}
func (UnimplementedMsgServer) CreateBridge(context.Context, *MsgCreateBridge) (*MsgCreateBridgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateBridge not implemented")
}
func (UnimplementedMsgServer) ProposeOutput(context.Context, *MsgProposeOutput) (*MsgProposeOutputResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProposeOutput not implemented")
}
func (UnimplementedMsgServer) DeleteOutput(context.Context, *MsgDeleteOutput) (*MsgDeleteOutputResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOutput not implemented")
}
func (UnimplementedMsgServer) InitiateTokenDeposit(context.Context, *MsgInitiateTokenDeposit) (*MsgInitiateTokenDepositResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitiateTokenDeposit not implemented")
}
func (UnimplementedMsgServer) FinalizeTokenWithdrawal(context.Context, *MsgFinalizeTokenWithdrawal) (*MsgFinalizeTokenWithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinalizeTokenWithdrawal not implemented")
}
func (UnimplementedMsgServer) UpdateProposer(context.Context, *MsgUpdateProposer) (*MsgUpdateProposerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProposer not implemented")
}
func (UnimplementedMsgServer) UpdateChallenger(context.Context, *MsgUpdateChallenger) (*MsgUpdateChallengerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateChallenger not implemented")
}
func (UnimplementedMsgServer) UpdateBatchInfo(context.Context, *MsgUpdateBatchInfo) (*MsgUpdateBatchInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBatchInfo not implemented")
}
func (UnimplementedMsgServer) UpdateMetadata(context.Context, *MsgUpdateMetadata) (*MsgUpdateMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMetadata not implemented")
}
func (UnimplementedMsgServer) UpdateOracleConfig(context.Context, *MsgUpdateOracleConfig) (*MsgUpdateOracleConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOracleConfig not implemented")
}
func (UnimplementedMsgServer) UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
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

func _Msg_RecordBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRecordBatch)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RecordBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_RecordBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RecordBatch(ctx, req.(*MsgRecordBatch))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateBridge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateBridge)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateBridge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateBridge_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateBridge(ctx, req.(*MsgCreateBridge))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ProposeOutput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgProposeOutput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ProposeOutput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_ProposeOutput_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ProposeOutput(ctx, req.(*MsgProposeOutput))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_DeleteOutput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgDeleteOutput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DeleteOutput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_DeleteOutput_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DeleteOutput(ctx, req.(*MsgDeleteOutput))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_InitiateTokenDeposit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgInitiateTokenDeposit)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).InitiateTokenDeposit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_InitiateTokenDeposit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).InitiateTokenDeposit(ctx, req.(*MsgInitiateTokenDeposit))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_FinalizeTokenWithdrawal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFinalizeTokenWithdrawal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FinalizeTokenWithdrawal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_FinalizeTokenWithdrawal_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FinalizeTokenWithdrawal(ctx, req.(*MsgFinalizeTokenWithdrawal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateProposer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateProposer)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateProposer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateProposer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateProposer(ctx, req.(*MsgUpdateProposer))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateChallenger_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateChallenger)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateChallenger(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateChallenger_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateChallenger(ctx, req.(*MsgUpdateChallenger))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateBatchInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateBatchInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateBatchInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateBatchInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateBatchInfo(ctx, req.(*MsgUpdateBatchInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateMetadata)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateMetadata_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateMetadata(ctx, req.(*MsgUpdateMetadata))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateOracleConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateOracleConfig)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateOracleConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateOracleConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateOracleConfig(ctx, req.(*MsgUpdateOracleConfig))
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

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "opinit.ophost.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RecordBatch",
			Handler:    _Msg_RecordBatch_Handler,
		},
		{
			MethodName: "CreateBridge",
			Handler:    _Msg_CreateBridge_Handler,
		},
		{
			MethodName: "ProposeOutput",
			Handler:    _Msg_ProposeOutput_Handler,
		},
		{
			MethodName: "DeleteOutput",
			Handler:    _Msg_DeleteOutput_Handler,
		},
		{
			MethodName: "InitiateTokenDeposit",
			Handler:    _Msg_InitiateTokenDeposit_Handler,
		},
		{
			MethodName: "FinalizeTokenWithdrawal",
			Handler:    _Msg_FinalizeTokenWithdrawal_Handler,
		},
		{
			MethodName: "UpdateProposer",
			Handler:    _Msg_UpdateProposer_Handler,
		},
		{
			MethodName: "UpdateChallenger",
			Handler:    _Msg_UpdateChallenger_Handler,
		},
		{
			MethodName: "UpdateBatchInfo",
			Handler:    _Msg_UpdateBatchInfo_Handler,
		},
		{
			MethodName: "UpdateMetadata",
			Handler:    _Msg_UpdateMetadata_Handler,
		},
		{
			MethodName: "UpdateOracleConfig",
			Handler:    _Msg_UpdateOracleConfig_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "opinit/ophost/v1/tx.proto",
}
