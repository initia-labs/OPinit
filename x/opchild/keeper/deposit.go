package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) handleBridgeHook(ctx sdk.Context, data []byte, hookMaxGas uint64) (success bool, reason string) {
	if hookMaxGas == 0 {
		return false, "hook max gas is zero"
	}

	originGasMeter := ctx.GasMeter()
	gasForHook := originGasMeter.GasRemaining()
	if gasForHook > hookMaxGas {
		gasForHook = hookMaxGas
	}

	defer func() {
		if r := recover(); r != nil {
			reason = fmt.Sprintf("panic: %v", r)
		}

		const maxReasonLength = 128
		if len(reason) > maxReasonLength {
			reason = reason[:maxReasonLength] + "..."
		}

		originGasMeter.ConsumeGas(ctx.GasMeter().GasConsumedToLimit(), "bridge hook")
	}()

	// use new gas meter with the hook max gas limit
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(gasForHook))

	tx, err := k.txDecoder(data)
	if err != nil {
		reason = fmt.Sprintf("Failed to decode tx: %s", err)
		return
	}

	ctx, err = k.decorators(ctx, tx, false)
	if err != nil {
		reason = fmt.Sprintf("Failed to run AnteHandler: %s", err)
		return
	}

	// use cache context from here to avoid resetting sequencer number on failure
	cacheCtx, commit := ctx.CacheContext()
	for _, msg := range tx.GetMsgs() {
		handler := k.msgRouter.Handler(msg)
		if handler == nil {
			reason = fmt.Sprintf("Unrecognized Msg type: %s", sdk.MsgTypeURL(msg))
			return
		}

		res, err := handler(cacheCtx, msg)
		if err != nil {
			reason = fmt.Sprintf("Failed to execute Msg: %s", err)
			return
		}

		// emit events
		cacheCtx.EventManager().EmitEvents(res.GetEvents())
	}

	commit()
	success = true

	return
}

// safeDepositToken mint and send coins to the recipient. Rollback all state changes
// if the deposit is failed.
func (ms MsgServer) safeDepositToken(ctx context.Context, toAddr sdk.AccAddress, coins sdk.Coins) (success bool, reason string) {
	// if coin is zero, just create an account
	if coins.IsZero() {
		if !ms.authKeeper.HasAccount(ctx, toAddr) {
			newAcc := ms.authKeeper.NewAccountWithAddress(ctx, toAddr)
			ms.authKeeper.SetAccount(ctx, newAcc)
		}

		return true, ""
	}

	var err error
	defer func() {
		if r := recover(); r != nil {
			reason = fmt.Sprintf("panic: %v", r)
		}

		const maxReasonLength = 128
		if len(reason) > maxReasonLength {
			reason = reason[:maxReasonLength] + "..."
		}
	}()

	// use cache context to avoid relaying failure
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, commit := sdkCtx.CacheContext()

	// mint coins to the module account
	if err = ms.bankKeeper.MintCoins(cacheCtx, types.ModuleName, coins); err != nil {
		reason = fmt.Sprintf("failed to mint coins: %s", err)
		return
	}

	// transfer can be failed due to contract logics
	if err = ms.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, toAddr, coins); err != nil {
		reason = fmt.Sprintf("failed to send coins: %s", err)
		return
	}

	// write the changes only if the transfer is successful
	commit()
	success = true

	return
}
