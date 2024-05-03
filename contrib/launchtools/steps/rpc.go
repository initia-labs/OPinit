package steps

import (
	"github.com/initia-labs/OPinit/contrib/launchtools"
	launchutils "github.com/initia-labs/OPinit/contrib/launchtools/utils"
	"github.com/pkg/errors"
)

// InitializeRPCHelpers initializes the RPC helpers for L1 and L2 chains.
// Each rpc helper is an isolated context to each chain. Useful for querying chain state or broadcasting transactions.
// Note that the codec is shared between the two helpers. This means that regardless of the chain, the codec registry
// must have registered all required types beforehand for any codec-related function to work properly.
// - Assumes launchtools.Launcher.ClientContext() already registered all necessary protobuf types.
func InitializeRPCHelpers(input launchtools.Input) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		ctx.Logger().Info("initializing RPC helpers",
			"l1-rpc-url", input.L1Config.RPCURL,
			"l1-chain-id", input.L1Config.ChainID,
		)

		l1, err := launchutils.NewRPCHelper(
			ctx.Logger().With("module", "rpc-helper"),
			input.L1Config.RPCURL,
			input.L1Config.ChainID,
			ctx.ClientContext().Codec,
			ctx.ClientContext().InterfaceRegistry,
			ctx.ClientContext().TxConfig,
		)

		if err != nil {
			return errors.Wrapf(err, "failed to create RPC client for L1")
		}

		ctx.Logger().Info("initializing RPC helpers",
			"l2-rpc-url", "http://localhost:26657",
			"l2-chain-id", input.L2Config.ChainID,
		)

		l2, err := launchutils.NewRPCHelper(
			ctx.Logger().With("module", "rpc-helper"),
			"http://localhost:26657",
			input.L2Config.ChainID,
			ctx.ClientContext().Codec,
			ctx.ClientContext().InterfaceRegistry,
			ctx.ClientContext().TxConfig,
		)

		ctx.SetRPCHelpers(l1, l2)

		return nil
	}
}
