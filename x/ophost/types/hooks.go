package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BridgeHook interface {
	BridgeCreated(
		ctx sdk.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeChallengerUpdated(
		ctx sdk.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeProposerUpdated(
		ctx sdk.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
}

type BridgeHooks []BridgeHook

func NewBridgeHooks(hooks ...BridgeHook) BridgeHooks {
	return hooks
}

var _ BridgeHook = BridgeHooks{}

func (hooks BridgeHooks) BridgeCreated(
	ctx sdk.Context,
	bridgeId uint64,
	bridgeConfig BridgeConfig,
) error {
	for _, h := range hooks {
		if err := h.BridgeCreated(ctx, bridgeId, bridgeConfig); err != nil {
			return err
		}
	}

	return nil
}

func (hooks BridgeHooks) BridgeChallengerUpdated(
	ctx sdk.Context,
	bridgeId uint64,
	bridgeConfig BridgeConfig,
) error {
	for _, h := range hooks {
		if err := h.BridgeChallengerUpdated(ctx, bridgeId, bridgeConfig); err != nil {
			return err
		}
	}

	return nil
}

func (hooks BridgeHooks) BridgeProposerUpdated(
	ctx sdk.Context,
	bridgeId uint64,
	bridgeConfig BridgeConfig,
) error {
	for _, h := range hooks {
		if err := h.BridgeProposerUpdated(ctx, bridgeId, bridgeConfig); err != nil {
			return err
		}
	}

	return nil
}
