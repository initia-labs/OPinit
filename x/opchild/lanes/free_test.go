package lanes_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	protov2 "google.golang.org/protobuf/proto"

	"github.com/initia-labs/OPinit/x/opchild/lanes"
)

var _ lanes.FeeWhitelistKeeper = MockFeeWhitelistKeeper{}

type MockFeeWhitelistKeeper struct {
	whitelist []string
}

// FeeWhitelist implements lanes.FeeWhitelistKeeper.
func (m MockFeeWhitelistKeeper) FeeWhitelist(ctx context.Context) ([]string, error) {
	return m.whitelist, nil
}

func Test_FreeLaneMatchHandler(t *testing.T) {
	ctx := sdk.NewContext(nil, types.Header{}, false, log.NewNopLogger())
	ac := address.NewBech32Codec("init")

	addr1, err := ac.BytesToString([]byte{0, 1, 2, 3})
	require.NoError(t, err)

	addr2, err := ac.BytesToString([]byte{3, 4, 5, 6})
	require.NoError(t, err)

	fwk := MockFeeWhitelistKeeper{
		whitelist: []string{addr1, addr2},
	}

	handler := lanes.NewFreeLaneMatchHandler(ac, fwk).MatchHandler()

	require.True(t, handler(ctx, MockTx{
		feePayer: []byte{0, 1, 2, 3},
	}))
	require.False(t, handler(ctx, MockTx{
		feePayer: []byte{0, 1, 2, 4},
	}))
	require.True(t, handler(ctx, MockTx{
		feePayer: []byte{3, 4, 5, 6},
	}))
}

var _ sdk.Tx = MockTx{}
var _ sdk.FeeTx = &MockTx{}

type MockTx struct {
	msgs     []sdk.Msg
	feePayer []byte
}

func (tx MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (tx MockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (tx MockTx) GetGas() uint64 {
	return 0
}

func (tx MockTx) GetFee() sdk.Coins {
	return nil
}

func (tx MockTx) FeePayer() []byte {
	return tx.feePayer
}
func (tx MockTx) FeeGranter() []byte {
	return nil
}
