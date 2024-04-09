// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: opinit/opchild/v1/types.proto

package types

import (
	fmt "fmt"
	types2 "github.com/cometbft/cometbft/abci/types"
	_ "github.com/cosmos/cosmos-proto"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params defines the set of opchild parameters.
type Params struct {
	// max_validators is the maximum number of validators.
	MaxValidators uint32 `protobuf:"varint,1,opt,name=max_validators,json=maxValidators,proto3" json:"max_validators,omitempty" yaml:"max_validators"`
	// historical_entries is the number of historical entries to persist.
	HistoricalEntries uint32                                      `protobuf:"varint,2,opt,name=historical_entries,json=historicalEntries,proto3" json:"historical_entries,omitempty" yaml:"historical_entries"`
	MinGasPrices      github_com_cosmos_cosmos_sdk_types.DecCoins `protobuf:"bytes,3,rep,name=min_gas_prices,json=minGasPrices,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.DecCoins" json:"min_gas_prices" yaml:"min_gas_price"`
	// the account address of bridge executor who can execute permissioned bridge
	// messages.
	BridgeExecutor string `protobuf:"bytes,4,opt,name=bridge_executor,json=bridgeExecutor,proto3" json:"bridge_executor,omitempty" yaml:"bridge_executor"`
	HostChainId    string `protobuf:"bytes,5,opt,name=host_chain_id,json=hostChainId,proto3" json:"host_chain_id,omitempty"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc6df244b706d68, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

// Validator defines a validator, together with the total amount of the
// Validator's bond shares and their exchange rate to coins. Slashing results in
// a decrease in the exchange rate, allowing correct calculation of future
// undelegations without iterating over delegators. When coins are delegated to
// this validator, the validator is credited with a delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonded shares
// multiplied by exchange rate.
type Validator struct {
	Moniker string `protobuf:"bytes,1,opt,name=moniker,proto3" json:"moniker,omitempty" yaml:"moniker"`
	// operator_address defines the address of the validator's operator;
	// bech encoded in JSON.
	OperatorAddress string `protobuf:"bytes,2,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty" yaml:"operator_address"`
	// consensus_pubkey is the consensus public key of the validator,
	// as a Protobuf Any.
	ConsensusPubkey *types1.Any `protobuf:"bytes,3,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty" yaml:"consensus_pubkey"`
	ConsPower       int64       `protobuf:"varint,4,opt,name=cons_power,json=consPower,proto3" json:"cons_power,omitempty" yaml:"cons_power"`
}

func (m *Validator) Reset()      { *m = Validator{} }
func (*Validator) ProtoMessage() {}
func (*Validator) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc6df244b706d68, []int{1}
}
func (m *Validator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Validator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Validator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Validator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Validator.Merge(m, src)
}
func (m *Validator) XXX_Size() int {
	return m.Size()
}
func (m *Validator) XXX_DiscardUnknown() {
	xxx_messageInfo_Validator.DiscardUnknown(m)
}

var xxx_messageInfo_Validator proto.InternalMessageInfo

// ValidatorUpdates defines an array of abci.ValidatorUpdate objects.
// TODO: explore moving this to proto/cosmos/base to separate modules
// from tendermint dependence
type ValidatorUpdates struct {
	Updates []types2.ValidatorUpdate `protobuf:"bytes,1,rep,name=updates,proto3" json:"updates"`
}

func (m *ValidatorUpdates) Reset()         { *m = ValidatorUpdates{} }
func (m *ValidatorUpdates) String() string { return proto.CompactTextString(m) }
func (*ValidatorUpdates) ProtoMessage()    {}
func (*ValidatorUpdates) Descriptor() ([]byte, []int) {
	return fileDescriptor_2cc6df244b706d68, []int{2}
}
func (m *ValidatorUpdates) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorUpdates) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorUpdates.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorUpdates) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorUpdates.Merge(m, src)
}
func (m *ValidatorUpdates) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorUpdates) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorUpdates.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorUpdates proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Params)(nil), "opinit.opchild.v1.Params")
	proto.RegisterType((*Validator)(nil), "opinit.opchild.v1.Validator")
	proto.RegisterType((*ValidatorUpdates)(nil), "opinit.opchild.v1.ValidatorUpdates")
}

func init() { proto.RegisterFile("opinit/opchild/v1/types.proto", fileDescriptor_2cc6df244b706d68) }

var fileDescriptor_2cc6df244b706d68 = []byte{
	// 727 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x54, 0x41, 0x4f, 0xdb, 0x48,
	0x14, 0xb6, 0x09, 0x0b, 0x8a, 0x81, 0x40, 0xac, 0xb0, 0x1b, 0x60, 0xb1, 0x23, 0x9f, 0x22, 0x76,
	0x63, 0x2b, 0xb0, 0xd2, 0x4a, 0x39, 0x2d, 0x66, 0xd9, 0x55, 0x45, 0x25, 0x22, 0x57, 0xed, 0xa1,
	0x17, 0x77, 0x6c, 0x4f, 0x9d, 0x11, 0xf1, 0x8c, 0xe5, 0x99, 0xa4, 0xc9, 0x3f, 0xa8, 0x7a, 0x69,
	0x8f, 0x3d, 0x72, 0x44, 0x3d, 0x71, 0xe8, 0xad, 0x7f, 0x00, 0x55, 0xaa, 0x84, 0x7a, 0xea, 0xc9,
	0x6d, 0xc3, 0x81, 0x9e, 0xf3, 0x0b, 0x2a, 0x7b, 0x26, 0x09, 0x4d, 0x2f, 0xc9, 0xbc, 0xf7, 0x7d,
	0xef, 0xcd, 0xf8, 0x7b, 0xdf, 0x8c, 0xb2, 0x4b, 0x62, 0x84, 0x11, 0xb3, 0x48, 0xec, 0x77, 0x50,
	0x37, 0xb0, 0xfa, 0x4d, 0x8b, 0x0d, 0x63, 0x48, 0xcd, 0x38, 0x21, 0x8c, 0xa8, 0x65, 0x0e, 0x9b,
	0x02, 0x36, 0xfb, 0xcd, 0xed, 0x32, 0x88, 0x10, 0x26, 0x56, 0xfe, 0xcb, 0x59, 0xdb, 0x9a, 0x4f,
	0x68, 0x44, 0xa8, 0xe5, 0x01, 0x0a, 0xad, 0x7e, 0xd3, 0x83, 0x0c, 0x34, 0x2d, 0x9f, 0x20, 0x2c,
	0xf0, 0x2d, 0x8e, 0xbb, 0x79, 0x64, 0xf1, 0x40, 0x40, 0x95, 0x90, 0x84, 0x84, 0xe7, 0xb3, 0xd5,
	0xa4, 0x20, 0x24, 0x24, 0xec, 0x42, 0x2b, 0x8f, 0xbc, 0xde, 0x53, 0x0b, 0xe0, 0xa1, 0x80, 0x76,
	0x18, 0xc4, 0x01, 0x4c, 0x22, 0x84, 0x99, 0x05, 0x3c, 0x1f, 0xdd, 0x3d, 0xae, 0xf1, 0xa1, 0xa0,
	0x2c, 0xb5, 0x41, 0x02, 0x22, 0xaa, 0xfe, 0xa3, 0x94, 0x22, 0x30, 0x70, 0xfb, 0xa0, 0x8b, 0x02,
	0xc0, 0x48, 0x42, 0xab, 0x72, 0x4d, 0xae, 0xaf, 0xd9, 0x5b, 0xe3, 0x54, 0xdf, 0x1c, 0x82, 0xa8,
	0xdb, 0x32, 0x7e, 0xc4, 0x0d, 0x67, 0x2d, 0x02, 0x83, 0x47, 0xd3, 0x58, 0xbd, 0xaf, 0xa8, 0x1d,
	0x44, 0x19, 0x49, 0x90, 0x0f, 0xba, 0x2e, 0xc4, 0x2c, 0x41, 0x90, 0x56, 0x17, 0xf2, 0x2e, 0xbb,
	0xe3, 0x54, 0xdf, 0xe2, 0x5d, 0x7e, 0xe6, 0x18, 0x4e, 0x79, 0x96, 0x3c, 0xe6, 0x39, 0xf5, 0xa5,
	0xac, 0x94, 0x22, 0x84, 0xdd, 0x10, 0x64, 0x3a, 0x20, 0x1f, 0xd2, 0x6a, 0xa1, 0x56, 0xa8, 0xaf,
	0xec, 0xff, 0x6e, 0x0a, 0x41, 0x32, 0xf5, 0x4c, 0xa1, 0x9e, 0xf9, 0x2f, 0xf4, 0x8f, 0x08, 0xc2,
	0xf6, 0xc9, 0x55, 0xaa, 0x4b, 0xe3, 0x54, 0xaf, 0x88, 0x23, 0xdf, 0xed, 0x60, 0xbc, 0xf9, 0xac,
	0xff, 0x11, 0x22, 0xd6, 0xe9, 0x79, 0xa6, 0x4f, 0x22, 0x21, 0xac, 0xf8, 0x6b, 0xd0, 0xe0, 0x4c,
	0x68, 0x23, 0x7a, 0x51, 0x67, 0x35, 0x42, 0xf8, 0x7f, 0x40, 0xdb, 0xf9, 0xf6, 0xea, 0x13, 0x65,
	0xdd, 0x4b, 0x50, 0x10, 0x42, 0x17, 0x0e, 0xa0, 0xdf, 0x63, 0x24, 0xa9, 0x2e, 0xd6, 0xe4, 0x7a,
	0xd1, 0xfe, 0x7b, 0x9c, 0xea, 0xbf, 0xf2, 0xfd, 0xe6, 0x08, 0xc6, 0xc7, 0xb7, 0x8d, 0x8a, 0x38,
	0xee, 0x61, 0x10, 0x24, 0x90, 0xd2, 0x07, 0x2c, 0x41, 0x38, 0xbc, 0xb8, 0xbd, 0xdc, 0x93, 0x9d,
	0x12, 0xa7, 0x1f, 0x0b, 0xb6, 0x6a, 0x28, 0x6b, 0x1d, 0x42, 0x99, 0xeb, 0x77, 0x00, 0xc2, 0x2e,
	0x0a, 0xaa, 0xbf, 0x64, 0xfd, 0x9d, 0x95, 0x2c, 0x79, 0x94, 0xe5, 0xee, 0x05, 0xad, 0x9d, 0xd7,
	0xe7, 0xba, 0xf4, 0xed, 0x5c, 0x97, 0x5f, 0xdc, 0x5e, 0xee, 0x95, 0x26, 0x36, 0xe4, 0x43, 0x34,
	0xde, 0x2d, 0x28, 0xc5, 0xe9, 0x44, 0xd4, 0x3f, 0x95, 0xe5, 0x88, 0x60, 0x74, 0x06, 0x93, 0x7c,
	0x96, 0x45, 0x5b, 0x1d, 0xa7, 0x7a, 0x49, 0x08, 0xc3, 0x01, 0xc3, 0x99, 0x50, 0xd4, 0xff, 0x94,
	0x0d, 0x12, 0xc3, 0x24, 0xab, 0x74, 0x01, 0x3f, 0x6b, 0x3e, 0xbc, 0xa2, 0xbd, 0x33, 0x4e, 0xf5,
	0xdf, 0x78, 0xd9, 0x3c, 0xc3, 0x70, 0xd6, 0x27, 0x29, 0xf1, 0x7d, 0x2a, 0x53, 0x36, 0x7c, 0x82,
	0x29, 0xc4, 0xb4, 0x47, 0xdd, 0xb8, 0xe7, 0x9d, 0xc1, 0x61, 0xb5, 0x50, 0x93, 0xeb, 0x2b, 0xfb,
	0x15, 0x93, 0xdb, 0xd4, 0x9c, 0xd8, 0xd4, 0x3c, 0xc4, 0x43, 0xfb, 0x60, 0xd6, 0x7d, 0xbe, 0xce,
	0x78, 0x3f, 0x93, 0xcf, 0x4f, 0x86, 0x31, 0x23, 0x66, 0xbb, 0xe7, 0x9d, 0xc0, 0xa1, 0xb3, 0x3e,
	0xa5, 0xb6, 0x73, 0xa6, 0xfa, 0x97, 0xa2, 0x64, 0x29, 0x37, 0x26, 0xcf, 0x20, 0x9f, 0x4b, 0xc1,
	0xde, 0x1c, 0xa7, 0x7a, 0x79, 0xd6, 0x99, 0x63, 0x86, 0x53, 0xcc, 0x82, 0x76, 0xb6, 0x6e, 0xad,
	0x3e, 0x3f, 0xd7, 0x25, 0x21, 0xa8, 0x64, 0xb8, 0xca, 0xc6, 0x54, 0xbc, 0x87, 0x71, 0x00, 0x18,
	0xa4, 0xea, 0xb1, 0xb2, 0xdc, 0xe3, 0xcb, 0xaa, 0x9c, 0xdb, 0xaf, 0x66, 0xce, 0x2e, 0x94, 0x99,
	0x5d, 0x28, 0x73, 0xae, 0xc6, 0x2e, 0x66, 0x16, 0xe4, 0x03, 0x9e, 0xd4, 0xb6, 0x16, 0xb3, 0x0d,
	0xec, 0xd3, 0xab, 0xaf, 0x9a, 0x74, 0x31, 0xd2, 0xe4, 0xab, 0x91, 0x26, 0x5f, 0x8f, 0x34, 0xf9,
	0xcb, 0x48, 0x93, 0x5f, 0xdd, 0x68, 0xd2, 0xf5, 0x8d, 0x26, 0x7d, 0xba, 0xd1, 0xa4, 0xc7, 0x8d,
	0x3b, 0xf6, 0xcc, 0x1e, 0x12, 0x04, 0x1a, 0x5d, 0xe0, 0x51, 0xeb, 0xb4, 0x9d, 0xbf, 0x3a, 0x83,
	0xe9, 0xbb, 0x93, 0x3b, 0xd5, 0x5b, 0xca, 0x95, 0x3c, 0xf8, 0x1e, 0x00, 0x00, 0xff, 0xff, 0x47,
	0xb5, 0x17, 0xc9, 0x96, 0x04, 0x00, 0x00,
}

func (this *Params) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxValidators != that1.MaxValidators {
		return false
	}
	if this.HistoricalEntries != that1.HistoricalEntries {
		return false
	}
	if len(this.MinGasPrices) != len(that1.MinGasPrices) {
		return false
	}
	for i := range this.MinGasPrices {
		if !this.MinGasPrices[i].Equal(&that1.MinGasPrices[i]) {
			return false
		}
	}
	if this.BridgeExecutor != that1.BridgeExecutor {
		return false
	}
	if this.HostChainId != that1.HostChainId {
		return false
	}
	return true
}
func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.HostChainId) > 0 {
		i -= len(m.HostChainId)
		copy(dAtA[i:], m.HostChainId)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.HostChainId)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.BridgeExecutor) > 0 {
		i -= len(m.BridgeExecutor)
		copy(dAtA[i:], m.BridgeExecutor)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.BridgeExecutor)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.MinGasPrices) > 0 {
		for iNdEx := len(m.MinGasPrices) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.MinGasPrices[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTypes(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.HistoricalEntries != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.HistoricalEntries))
		i--
		dAtA[i] = 0x10
	}
	if m.MaxValidators != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.MaxValidators))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Validator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Validator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Validator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ConsPower != 0 {
		i = encodeVarintTypes(dAtA, i, uint64(m.ConsPower))
		i--
		dAtA[i] = 0x20
	}
	if m.ConsensusPubkey != nil {
		{
			size, err := m.ConsensusPubkey.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTypes(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if len(m.OperatorAddress) > 0 {
		i -= len(m.OperatorAddress)
		copy(dAtA[i:], m.OperatorAddress)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.OperatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Moniker) > 0 {
		i -= len(m.Moniker)
		copy(dAtA[i:], m.Moniker)
		i = encodeVarintTypes(dAtA, i, uint64(len(m.Moniker)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ValidatorUpdates) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorUpdates) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorUpdates) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Updates) > 0 {
		for iNdEx := len(m.Updates) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Updates[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTypes(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintTypes(dAtA []byte, offset int, v uint64) int {
	offset -= sovTypes(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MaxValidators != 0 {
		n += 1 + sovTypes(uint64(m.MaxValidators))
	}
	if m.HistoricalEntries != 0 {
		n += 1 + sovTypes(uint64(m.HistoricalEntries))
	}
	if len(m.MinGasPrices) > 0 {
		for _, e := range m.MinGasPrices {
			l = e.Size()
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	l = len(m.BridgeExecutor)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.HostChainId)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	return n
}

func (m *Validator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Moniker)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	l = len(m.OperatorAddress)
	if l > 0 {
		n += 1 + l + sovTypes(uint64(l))
	}
	if m.ConsensusPubkey != nil {
		l = m.ConsensusPubkey.Size()
		n += 1 + l + sovTypes(uint64(l))
	}
	if m.ConsPower != 0 {
		n += 1 + sovTypes(uint64(m.ConsPower))
	}
	return n
}

func (m *ValidatorUpdates) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Updates) > 0 {
		for _, e := range m.Updates {
			l = e.Size()
			n += 1 + l + sovTypes(uint64(l))
		}
	}
	return n
}

func sovTypes(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTypes(x uint64) (n int) {
	return sovTypes(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxValidators", wireType)
			}
			m.MaxValidators = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxValidators |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HistoricalEntries", wireType)
			}
			m.HistoricalEntries = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HistoricalEntries |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinGasPrices", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MinGasPrices = append(m.MinGasPrices, types.DecCoin{})
			if err := m.MinGasPrices[len(m.MinGasPrices)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BridgeExecutor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BridgeExecutor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HostChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HostChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Validator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Validator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Validator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Moniker", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Moniker = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OperatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsensusPubkey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ConsensusPubkey == nil {
				m.ConsensusPubkey = &types1.Any{}
			}
			if err := m.ConsensusPubkey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsPower", wireType)
			}
			m.ConsPower = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ConsPower |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValidatorUpdates) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValidatorUpdates: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorUpdates: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Updates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTypes
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTypes
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Updates = append(m.Updates, types2.ValidatorUpdate{})
			if err := m.Updates[len(m.Updates)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTypes(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTypes
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTypes(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTypes
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTypes
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTypes
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTypes
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTypes
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTypes        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTypes          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTypes = fmt.Errorf("proto: unexpected end of group")
)
