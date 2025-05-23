package types

import (
	"bytes"
	"testing"
	time "time"

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

func Test_MsgUpdateOracleConfig(t *testing.T) {
	ac := address.NewBech32Codec("init")
	addr, err := ac.BytesToString(bytes.Repeat([]byte{1}, 20))
	require.NoError(t, err)

	// Valid input
	validMsg := NewMsgUpdateOracleConfig(addr, 123, true)
	require.NoError(t, validMsg.Validate(ac))

	// Empty submitter
	invalidMsg := NewMsgUpdateOracleConfig("", 123, false)
	require.Error(t, invalidMsg.Validate(ac))
}

func Test_MsgUpdateFinalizationPeriod(t *testing.T) {
	ac := address.NewBech32Codec("init")
	addr, err := ac.BytesToString(bytes.Repeat([]byte{1}, 20))
	require.NoError(t, err)

	// Valid input
	validMsg := NewMsgUpdateFinalizationPeriod(addr, 1, time.Second*10)
	require.NoError(t, validMsg.Validate(ac))

	// Empty submitter
	invalidMsg := NewMsgUpdateFinalizationPeriod("", 1, time.Second*10)
	require.Error(t, invalidMsg.Validate(ac))

	// Invalid bridge id
	invalidMsg = NewMsgUpdateFinalizationPeriod(addr, 0, time.Second*10)
	require.Error(t, invalidMsg.Validate(ac))

	// Invalid finalization period
	invalidMsg = NewMsgUpdateFinalizationPeriod(addr, 1, time.Second*0)
	require.Error(t, invalidMsg.Validate(ac))
}
