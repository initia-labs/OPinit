package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
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
