package lanes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	blockbase "github.com/skip-mev/block-sdk/v2/block/base"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// SystemLaneMatchHandler returns the default match handler for the system lane.
func SystemLaneMatchHandler() blockbase.MatchHandler {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		for _, msg := range tx.GetMsgs() {
			switch msg := msg.(type) {
			case *clienttypes.MsgUpdateClient:
			case *types.MsgUpdateOracle:
			case *types.MsgRelayOracleData:
			case *authz.MsgExec:
				msgs, err := msg.GetMessages()
				if err != nil || len(msgs) != 1 {
					return false
				}
				switch msgs[0].(type) {
				case *types.MsgUpdateOracle, *types.MsgRelayOracleData:
				default:
					return false
				}
			default:
				return false
			}
		}

		return true
	}
}
