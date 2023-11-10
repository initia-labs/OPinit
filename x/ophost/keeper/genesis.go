package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) (res []abci.ValidatorUpdate) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, bridge := range data.Bridges {
		bridgeId := bridge.BridgeId
		if err := k.SetBridgeConfig(ctx, bridgeId, bridge.BridgeConfig); err != nil {
			panic(err)
		}

		k.SetNextL1Sequence(ctx, bridgeId, bridge.NextL1Sequence)

		for _, proposal := range bridge.Proposals {
			if err := k.SetOutputProposal(ctx, bridgeId, proposal.OutputIndex, proposal.OutputProposal); err != nil {
				panic(err)
			}
		}

		k.SetNextOutputIndex(ctx, bridgeId, bridge.NextOutputIndex)

		for _, provenWithdrawal := range bridge.ProvenWithdrawals {
			withdrawalHash := [32]byte{}
			copy(withdrawalHash[:], provenWithdrawal)
			k.RecordProvenWithdrawal(ctx, bridgeId, withdrawalHash)
		}

		for _, tokenPair := range bridge.TokenPairs {
			k.SetTokenPair(ctx, bridgeId, tokenPair.L2Denom, tokenPair.L1Denom)
		}
	}

	k.SetNextBridgeId(ctx, data.NextBridgeId)

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {

	var bridges []types.Bridge
	k.IterateBridgeConfig(ctx, func(bridgeId uint64, bridgeConfig types.BridgeConfig) bool {
		nextL1Sequence := k.GetNextL1Sequence(ctx, bridgeId)
		nextOutputIndex := k.GetNextOutputIndex(ctx, bridgeId)

		var proposals []types.WrappedOutput
		if err := k.IterateOutputProposals(ctx, bridgeId, func(bridgeId, outputIndex uint64, output types.Output) bool {
			proposals = append(proposals, types.WrappedOutput{
				OutputIndex:    outputIndex,
				OutputProposal: output,
			})

			return false
		}); err != nil {
			panic(err)
		}

		var provenWithdrawals [][]byte
		if err := k.IterateProvenWithdrawals(ctx, bridgeId, func(bridgeId uint64, withdrawalHash [32]byte) bool {
			provenWithdrawals = append(provenWithdrawals, withdrawalHash[:])
			return false
		}); err != nil {
			panic(err)
		}

		var tokenPairs []types.TokenPair
		if err := k.IterateTokenPair(ctx, bridgeId, func(bridgeId uint64, tokenPair types.TokenPair) bool {
			tokenPairs = append(tokenPairs, tokenPair)

			return false
		}); err != nil {
			panic(err)
		}

		bridges = append(bridges, types.Bridge{
			BridgeId:          bridgeId,
			NextL1Sequence:    nextL1Sequence,
			NextOutputIndex:   nextOutputIndex,
			BridgeConfig:      bridgeConfig,
			TokenPairs:        tokenPairs,
			ProvenWithdrawals: provenWithdrawals,
			Proposals:         proposals,
		})

		return false
	})

	nextBridgeId := k.GetNextBridgeId(ctx)

	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		Bridges:      bridges,
		NextBridgeId: nextBridgeId,
	}
}
