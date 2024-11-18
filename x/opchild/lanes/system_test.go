package lanes_test

import (
	"testing"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/stretchr/testify/require"

	"github.com/initia-labs/OPinit/x/opchild/lanes"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
)

func Test_SystemLaneMatchHandler(t *testing.T) {
	ctx := sdk.NewContext(nil, types.Header{}, false, log.NewNopLogger())

	handler := lanes.SystemLaneMatchHandler()

	// 1 system message
	require.True(t, handler(ctx, MockTx{
		msgs: []sdk.Msg{
			&opchildtypes.MsgUpdateOracle{},
		},
	}))

	// 2 system messages
	require.False(t, handler(ctx, MockTx{
		msgs: []sdk.Msg{
			&opchildtypes.MsgUpdateOracle{},
			&opchildtypes.MsgUpdateOracle{},
		},
	}))

	// 1 non-system message
	require.False(t, handler(ctx, MockTx{
		msgs: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	}))

	authzMsg := authz.NewMsgExec(nil, []sdk.Msg{
		&opchildtypes.MsgUpdateOracle{},
	})

	// 1 system message and authz.MsgExec
	require.True(t, handler(ctx, MockTx{
		msgs: []sdk.Msg{
			&authzMsg,
		},
	}))

	authzMsg = authz.NewMsgExec(nil, []sdk.Msg{
		&banktypes.MsgSend{},
	})

	// authz.MsgExec with non-system message
	require.False(t, handler(ctx, MockTx{
		msgs: []sdk.Msg{
			&authzMsg,
		},
	}))
}
