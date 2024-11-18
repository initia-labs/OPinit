package lanes

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	blockbase "github.com/skip-mev/block-sdk/v2/block/base"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// SystemLaneMatchHandler returns the default match handler for the system lane.
func SystemLaneMatchHandler() blockbase.MatchHandler {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		if len(tx.GetMsgs()) != 1 {
			return false
		}

		for _, msg := range tx.GetMsgs() {
			switch msg := msg.(type) {
			case *types.MsgUpdateOracle:
			case *authz.MsgExec:
				msgs, err := msg.GetMessages()
				if err != nil || len(msgs) != 1 {
					return false
				} else if _, ok := msgs[0].(*types.MsgUpdateOracle); !ok {
					return false
				}
			default:
				return false
			}
		}

		return true
	}
}
