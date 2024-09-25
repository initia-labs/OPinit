package types

import "context"

type BridgeHook interface {
	BridgeCreated(
		ctx context.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeChallengerUpdated(
		ctx context.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeProposerUpdated(
		ctx context.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeBatchInfoUpdated(
		ctx context.Context,
		bridgeId uint64,
		bridgeConfig BridgeConfig,
	) error
	BridgeMetadataUpdated(
		ctx context.Context,
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
	ctx context.Context,
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
	ctx context.Context,
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
	ctx context.Context,
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

func (hooks BridgeHooks) BridgeBatchInfoUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig BridgeConfig,
) error {
	for _, h := range hooks {
		if err := h.BridgeBatchInfoUpdated(ctx, bridgeId, bridgeConfig); err != nil {
			return err
		}
	}

	return nil
}

func (hooks BridgeHooks) BridgeMetadataUpdated(
	ctx context.Context,
	bridgeId uint64,
	bridgeConfig BridgeConfig,
) error {
	for _, h := range hooks {
		if err := h.BridgeMetadataUpdated(ctx, bridgeId, bridgeConfig); err != nil {
			return err
		}
	}

	return nil
}
