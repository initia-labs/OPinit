package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/ophost/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, bridge := range data.Bridges {
		bridgeId := bridge.BridgeId
		if err := k.SetBridgeConfig(ctx, bridgeId, bridge.BridgeConfig); err != nil {
			panic(err)
		}

		if err := k.SetNextL1Sequence(ctx, bridgeId, bridge.NextL1Sequence); err != nil {
			panic(err)
		}

		for _, proposal := range bridge.Proposals {
			if err := k.SetOutputProposal(ctx, bridgeId, proposal.OutputIndex, proposal.OutputProposal); err != nil {
				panic(err)
			}
		}

		if err := k.SetNextOutputIndex(ctx, bridgeId, bridge.NextOutputIndex); err != nil {
			panic(err)
		}

		for _, provenWithdrawal := range bridge.ProvenWithdrawals {
			withdrawalHash := [32]byte{}
			copy(withdrawalHash[:], provenWithdrawal)
			if err := k.RecordProvenWithdrawal(ctx, bridgeId, withdrawalHash); err != nil {
				panic(err)
			}
		}

		for _, tokenPair := range bridge.TokenPairs {
			if err := k.SetTokenPair(ctx, bridgeId, tokenPair.L2Denom, tokenPair.L1Denom); err != nil {
				panic(err)
			}
		}

		for _, batchInfo := range bridge.BatchInfos {
			if err := k.SetBatchInfo(ctx, bridgeId, batchInfo.BatchInfo, batchInfo.Output); err != nil {
				panic(err)
			}
		}
	}

	if err := k.SetNextBridgeId(ctx, data.NextBridgeId); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {

	var bridges []types.Bridge
	err := k.IterateBridgeConfig(ctx, func(bridgeId uint64, bridgeConfig types.BridgeConfig) (stop bool, err error) {
		nextL1Sequence, err := k.GetNextL1Sequence(ctx, bridgeId)
		if err != nil {
			return true, err
		}

		nextOutputIndex, err := k.GetNextOutputIndex(ctx, bridgeId)
		if err != nil {
			return true, err
		}

		var proposals []types.WrappedOutput
		if err := k.IterateOutputProposals(ctx, bridgeId, func(key collections.Pair[uint64, uint64], output types.Output) (stop bool, err error) {
			proposals = append(proposals, types.WrappedOutput{
				OutputIndex:    key.K2(),
				OutputProposal: output,
			})

			return false, nil
		}); err != nil {
			return true, err
		}

		var provenWithdrawals [][]byte
		if err := k.IterateProvenWithdrawals(ctx, bridgeId, func(bridgeId uint64, withdrawalHash [32]byte) (bool, error) {
			provenWithdrawals = append(provenWithdrawals, withdrawalHash[:])
			return false, nil
		}); err != nil {
			return true, err
		}

		var tokenPairs []types.TokenPair
		if err := k.IterateTokenPair(ctx, bridgeId, func(bridgeId uint64, tokenPair types.TokenPair) (stop bool, err error) {
			tokenPairs = append(tokenPairs, tokenPair)
			return false, nil
		}); err != nil {
			return true, err
		}

		var batchInfos []types.BatchInfoWithOutput
		if err := k.IterateBatchInfos(ctx, bridgeId, func(key collections.Pair[uint64, uint64], batchInfo types.BatchInfoWithOutput) (stop bool, err error) {
			batchInfos = append(batchInfos, batchInfo)
			return false, nil
		}); err != nil {
			return true, err
		}

		bridges = append(bridges, types.Bridge{
			BridgeId:          bridgeId,
			NextL1Sequence:    nextL1Sequence,
			NextOutputIndex:   nextOutputIndex,
			BridgeConfig:      bridgeConfig,
			TokenPairs:        tokenPairs,
			ProvenWithdrawals: provenWithdrawals,
			Proposals:         proposals,
			BatchInfos:        batchInfos,
		})

		return false, nil
	})
	if err != nil {
		panic(err)
	}

	nextBridgeId, err := k.GetNextBridgeId(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		Bridges:      bridges,
		NextBridgeId: nextBridgeId,
	}
}
