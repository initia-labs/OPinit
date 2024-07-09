package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

type RedundantBridgeDecorator struct {
	ms keeper.MsgServer
}

func NewRedundantBridgeDecorator(k keeper.Keeper) RedundantBridgeDecorator {
	return RedundantBridgeDecorator{
		ms: keeper.NewMsgServerImpl(k),
	}
}

func (rbd RedundantBridgeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if (ctx.IsCheckTx() || ctx.IsReCheckTx()) && !simulate {
		redundancies := 0
		packetMsgs := 0
		for _, m := range tx.GetMsgs() {
			switch msg := m.(type) {
			case *types.MsgFinalizeTokenDeposit:
				response, err := rbd.ms.FinalizeTokenDeposit(ctx, msg)
				if err != nil {
					return ctx, err
				}
				if response.Result == types.NOOP {
					redundancies++
				}
				packetMsgs++
			}
		}

		if redundancies == packetMsgs && packetMsgs > 0 {
			return ctx, types.ErrRedundantTx
		}
	}
	return next(ctx, tx, simulate)
}
