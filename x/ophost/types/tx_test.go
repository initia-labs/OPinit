package types

import (
	"bytes"
	"testing"

	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgRecordBatch_Validate(t *testing.T) {
	ac := address.NewBech32Codec("init")
	addr, err := ac.BytesToString(bytes.Repeat([]byte{1}, 20))
	require.NoError(t, err)

	// Valid input
	validMsg := NewMsgRecordBatch(addr, 123, bytes.Repeat([]byte{1}, 32))
	require.NoError(t, validMsg.Validate(ac))

	// Empty submitter
	invalidMsg := NewMsgRecordBatch("", 123, bytes.Repeat([]byte{1}, 32))
	require.Error(t, invalidMsg.Validate(ac))

	// Empty batch bytes
	invalidMsg = NewMsgRecordBatch(addr, 123, nil)
	require.Error(t, invalidMsg.Validate(ac))
}

func TestMsgForceTokenWithdrawal(t *testing.T) {
	ac := address.NewBech32Codec("init")
	addr, err := ac.BytesToString(bytes.Repeat([]byte{1}, 20))
	require.NoError(t, err)

	// Valid input
	validMsg := NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.NoError(t, validMsg.Validate(ac))

	// Empty app hash
	invalidMsg := NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		nil,
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty sender
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, "", addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty receiver
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, "", sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty amount
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 0), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Invalid version
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	invalidMsg.Version = []byte{1, 2}
	require.Error(t, invalidMsg.Validate(ac))

	// Empty storage root
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), nil, bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty latest block hash
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), nil,
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty commitment proof
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		1, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))

	// Empty l2 sequence
	invalidMsg = NewMsgForceTokenWithdrawal(
		123, 456,
		0, addr, addr, sdk.NewInt64Coin("test", 100), v1.ProofOps{Ops: []v1.ProofOp{{}}},
		bytes.Repeat([]byte{1}, 32),
		v1.Proof{Total: 1, Index: 1, LeafHash: bytes.Repeat([]byte{1}, 32), Aunts: [][]byte{bytes.Repeat([]byte{2}, 32)}},
		byte(1), bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{1}, 32),
	)
	require.Error(t, invalidMsg.Validate(ac))
}
