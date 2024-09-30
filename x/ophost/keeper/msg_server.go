package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

type MsgServer struct {
	Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServerImpl return MsgServer instance
func NewMsgServerImpl(k Keeper) MsgServer {
	return MsgServer{k}
}

/////////////////////////////////////////////////////
// The messages for Batch Submitter

// RecordBatch implements a RecordBatch message handling
func (ms MsgServer) RecordBatch(ctx context.Context, req *types.MsgRecordBatch) (*types.MsgRecordBatchResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecordBatch,
			sdk.NewAttribute(types.AttributeKeySubmitter, req.Submitter),
		),
	)

	return &types.MsgRecordBatchResponse{}, nil
}

/////////////////////////////////////////////////////
// The messages for Bridge Creator

func (ms MsgServer) CreateBridge(ctx context.Context, req *types.MsgCreateBridge) (*types.MsgCreateBridgeResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	// registration fee check
	registrationFee := ms.RegistrationFee(ctx)
	if registrationFee.IsValid() {
		creator, err := ms.authKeeper.AddressCodec().StringToBytes(req.Creator)
		if err != nil {
			return nil, err
		}

		err = ms.communityPoolKeeper.FundCommunityPool(ctx, registrationFee, creator)
		if err != nil {
			return nil, err
		}
	}

	bridgeId, err := ms.IncreaseNextBridgeId(ctx)
	if err != nil {
		return nil, err
	}

	// store bridge config
	if err := ms.SetBridgeConfig(ctx, bridgeId, req.Config); err != nil {
		return nil, err
	}

	err = ms.SetBatchInfo(ctx, bridgeId, req.Config.BatchInfo, types.Output{})
	if err != nil {
		return nil, err
	}

	// create bridge account
	bridgeAcc := types.NewBridgeAccountWithAddress(types.BridgeAddress(bridgeId))
	bridgeAccI := (ms.authKeeper.NewAccount(ctx, bridgeAcc)) // set the account number
	ms.authKeeper.SetAccount(ctx, bridgeAccI)

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreateBridge,
		sdk.NewAttribute(types.AttributeKeyCreator, req.Creator),
		sdk.NewAttribute(types.AttributeKeyProposer, req.Config.Proposer),
		sdk.NewAttribute(types.AttributeKeyChallenger, req.Config.Challenger),
		sdk.NewAttribute(types.AttributeKeyBatchChainType, req.Config.BatchInfo.ChainType.StringWithoutPrefix()),
		sdk.NewAttribute(types.AttributeKeyBatchSubmitter, req.Config.BatchInfo.Submitter),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOracleEnabled, strconv.FormatBool(req.Config.OracleEnabled)),
	))

	if err := ms.bridgeHook.BridgeCreated(ctx, bridgeId, req.Config); err != nil {
		return nil, err
	}

	return &types.MsgCreateBridgeResponse{
		BridgeId: bridgeId,
	}, nil
}

func (ms MsgServer) ProposeOutput(ctx context.Context, req *types.MsgProposeOutput) (*types.MsgProposeOutputResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	proposer := req.Proposer
	bridgeId := req.BridgeId
	l2BlockNumber := req.L2BlockNumber
	outputRoot := req.OutputRoot

	bridgeConfig, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// permission check
	if proposer != bridgeConfig.Proposer {
		return nil, errors.ErrUnauthorized.Wrap("invalid proposer")
	}

	// fetch next output index
	outputIndex, err := ms.IncreaseNextOutputIndex(ctx, bridgeId)
	if err != nil {
		return nil, err
	}
	if outputIndex != req.OutputIndex {
		return nil, types.ErrInvalidOutputIndex.Wrapf("expected %d, got %d", outputIndex, req.OutputIndex)
	}

	// check this is first submission or not
	if outputIndex != 1 {
		lastOutputProposal, err := ms.GetOutputProposal(ctx, bridgeId, outputIndex-1)
		if err != nil {
			return nil, err
		}

		if l2BlockNumber <= lastOutputProposal.L2BlockNumber {
			return nil, types.ErrInvalidL2BlockNumber.Wrapf("last %d, got %d", lastOutputProposal.L2BlockNumber, l2BlockNumber)
		}
	}

	// store output proposal
	if err := ms.SetOutputProposal(ctx, bridgeId, outputIndex, types.Output{
		OutputRoot:    outputRoot,
		L1BlockNumber: uint64(sdkCtx.BlockHeight()), //nolint:gosec
		L1BlockTime:   sdkCtx.BlockTime(),
		L2BlockNumber: l2BlockNumber,
	}); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeProposeOutput,
		sdk.NewAttribute(types.AttributeKeyProposer, proposer),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyL2BlockNumber, strconv.FormatUint(l2BlockNumber, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputRoot, hex.EncodeToString(outputRoot)),
	))

	return &types.MsgProposeOutputResponse{}, nil
}

func (ms MsgServer) DeleteOutput(ctx context.Context, req *types.MsgDeleteOutput) (*types.MsgDeleteOutputResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	challenger := req.Challenger
	bridgeId := req.BridgeId
	outputIndex := req.OutputIndex

	bridgeConfig, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// permission check
	if bridgeConfig.Challenger != challenger {
		return nil, errors.ErrUnauthorized.Wrap("invalid challenger")
	}

	nextOutputIndex, err := ms.GetNextOutputIndex(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	if outputIndex >= nextOutputIndex {
		return nil, types.ErrInvalidOutputIndex
	}

	// delete output proposals in [outputIndex, nextOutputIndex) range
	for i := outputIndex; i < nextOutputIndex; i++ {
		if err := ms.DeleteOutputProposal(ctx, bridgeId, i); err != nil {
			return nil, err
		}
	}

	// rollback next output index to the deleted output index
	if err := ms.NextOutputIndexes.Set(ctx, bridgeId, outputIndex); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeleteOutput,
		sdk.NewAttribute(types.AttributeKeyChallenger, challenger),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
	))

	return &types.MsgDeleteOutputResponse{}, nil
}

func (ms MsgServer) InitiateTokenDeposit(ctx context.Context, req *types.MsgInitiateTokenDeposit) (*types.MsgInitiateTokenDepositResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	sender, err := ms.authKeeper.AddressCodec().StringToBytes(req.Sender)
	if err != nil {
		return nil, err
	}

	coin := req.Amount
	bridgeId := req.BridgeId
	l1Sequence, err := ms.IncreaseNextL1Sequence(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// transfer only positive amount
	if coin.IsPositive() {
		// send the funds to bridge address
		bridgeAddr := types.BridgeAddress(bridgeId)
		if err := ms.bankKeeper.SendCoins(ctx, sender, bridgeAddr, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}

	// record token pairs
	l2Denom := types.L2Denom(bridgeId, coin.Denom)
	if ok, err := ms.HasTokenPair(ctx, bridgeId, l2Denom); err != nil {
		return nil, err
	} else if !ok {
		if err := ms.SetTokenPair(ctx, bridgeId, l2Denom, coin.Denom); err != nil {
			return nil, err
		}
	}

	// emit events for bridge executor
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInitiateTokenDeposit,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyL1Sequence, strconv.FormatUint(l1Sequence, 10)),
		sdk.NewAttribute(types.AttributeKeyFrom, req.Sender),
		sdk.NewAttribute(types.AttributeKeyTo, req.To),
		sdk.NewAttribute(types.AttributeKeyL1Denom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyL2Denom, l2Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyData, hex.EncodeToString(req.Data)),
	))

	return &types.MsgInitiateTokenDepositResponse{
		Sequence: l1Sequence,
	}, nil
}

func (ms MsgServer) FinalizeTokenWithdrawal(ctx context.Context, req *types.MsgFinalizeTokenWithdrawal) (*types.MsgFinalizeTokenWithdrawalResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	receiver, err := ms.authKeeper.AddressCodec().StringToBytes(req.To)
	if err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	outputIndex := req.OutputIndex
	l2Sequence := req.Sequence
	amount := req.Amount.Amount
	denom := req.Amount.Denom

	if ok, err := ms.IsFinalized(ctx, bridgeId, outputIndex); err != nil {
		return nil, err
	} else if !ok {
		return nil, types.ErrNotFinalized
	}

	outputProposal, err := ms.GetOutputProposal(ctx, bridgeId, outputIndex)
	if err != nil {
		return nil, err
	}

	// validate output root generation
	outputRoot := types.GenerateOutputRoot(req.Version[0], req.StorageRoot, req.LastBlockHash)
	if !bytes.Equal(outputProposal.OutputRoot, outputRoot[:]) {
		return nil, types.ErrFailedToVerifyWithdrawal.Wrap("invalid output root")
	}

	// verify storage root can be generated from
	// withdrawal proofs and withdrawal tx data.
	withdrawalHash := types.GenerateWithdrawalHash(bridgeId, l2Sequence, req.From, req.To, denom, amount.Uint64())
	if ok, err := ms.HasProvenWithdrawal(ctx, bridgeId, withdrawalHash); err != nil {
		return nil, err
	} else if ok {
		return nil, types.ErrWithdrawalAlreadyFinalized
	}

	// should use the same node hash generation function `types.GenerateNodeHash`
	// to make this node hash generation deterministic with commutative property.
	rootHash := types.GenerateRootHashFromProofs(withdrawalHash, req.WithdrawalProofs)
	if !bytes.Equal(req.StorageRoot, rootHash[:]) {
		return nil, types.ErrFailedToVerifyWithdrawal.Wrap("invalid storage root proofs")
	}

	if err := ms.RecordProvenWithdrawal(ctx, bridgeId, withdrawalHash); err != nil {
		return nil, err
	}

	// transfer asset to a user from the bridge account
	bridgeAddr := types.BridgeAddress(bridgeId)
	if err := ms.bankKeeper.SendCoins(ctx, bridgeAddr, receiver, sdk.NewCoins(sdk.NewCoin(denom, amount))); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeFinalizeTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, strconv.FormatUint(l2Sequence, 10)),
		sdk.NewAttribute(types.AttributeKeyFrom, req.From),
		sdk.NewAttribute(types.AttributeKeyTo, req.To),
		sdk.NewAttribute(types.AttributeKeyL1Denom, denom),
		sdk.NewAttribute(types.AttributeKeyL2Denom, types.L2Denom(bridgeId, denom)),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))

	return &types.MsgFinalizeTokenWithdrawalResponse{}, nil
}

func (ms MsgServer) UpdateProposer(ctx context.Context, req *types.MsgUpdateProposer) (*types.MsgUpdateProposerResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// gov or current proposer can update proposer.
	if ms.authority != req.Authority && config.Proposer != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s or %s, got %s", ms.authority, config.Proposer, req.Authority)
	}

	config.Proposer = req.NewProposer
	if err := ms.Keeper.bridgeHook.BridgeProposerUpdated(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	finalizedOutputIndex, finalizedOutput, err := ms.GetLastFinalizedOutput(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateProposer,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyProposer, req.NewProposer),
		sdk.NewAttribute(types.AttributeKeyFinalizedOutputIndex, strconv.FormatUint(finalizedOutputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyFinalizedL2BlockNumber, strconv.FormatUint(finalizedOutput.L2BlockNumber, 10)),
	))

	return &types.MsgUpdateProposerResponse{
		OutputIndex:   finalizedOutputIndex,
		L2BlockNumber: finalizedOutput.L2BlockNumber,
	}, nil
}

func (ms MsgServer) UpdateChallenger(ctx context.Context, req *types.MsgUpdateChallenger) (*types.MsgUpdateChallengerResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// permission check

	if req.Authority != ms.authority && config.Challenger != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s or %s, but got %s", ms.authority, config.Challenger, req.Authority)
	}

	config.Challenger = req.Challenger
	if err := ms.Keeper.bridgeHook.BridgeChallengerUpdated(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}
	finalizedOutputIndex, finalizedOutput, err := ms.GetLastFinalizedOutput(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateChallenger,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyChallenger, config.Challenger),
		sdk.NewAttribute(types.AttributeKeyFinalizedOutputIndex, strconv.FormatUint(finalizedOutputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyFinalizedL2BlockNumber, strconv.FormatUint(finalizedOutput.L2BlockNumber, 10)),
	))

	return &types.MsgUpdateChallengerResponse{
		OutputIndex:   finalizedOutputIndex,
		L2BlockNumber: finalizedOutput.L2BlockNumber,
	}, nil
}

func (ms MsgServer) UpdateBatchInfo(ctx context.Context, req *types.MsgUpdateBatchInfo) (*types.MsgUpdateBatchInfoResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// gov or current proposer can update batch info.
	if ms.authority != req.Authority && config.Proposer != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s or %s, got %s", ms.authority, config.Proposer, req.Authority)
	}

	config.BatchInfo = req.NewBatchInfo
	if err := ms.Keeper.bridgeHook.BridgeBatchInfoUpdated(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	finalizedOutputIndex, finalizedOutput, err := ms.GetLastFinalizedOutput(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	err = ms.SetBatchInfo(ctx, bridgeId, req.NewBatchInfo, finalizedOutput)
	if err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateBatchInfo,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyBatchChainType, req.NewBatchInfo.ChainType.StringWithoutPrefix()),
		sdk.NewAttribute(types.AttributeKeyBatchSubmitter, req.NewBatchInfo.Submitter),
		sdk.NewAttribute(types.AttributeKeyFinalizedOutputIndex, strconv.FormatUint(finalizedOutputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyFinalizedL2BlockNumber, strconv.FormatUint(finalizedOutput.L2BlockNumber, 10)),
	))

	return &types.MsgUpdateBatchInfoResponse{
		OutputIndex:   finalizedOutputIndex,
		L2BlockNumber: finalizedOutput.L2BlockNumber,
	}, nil
}

func (ms MsgServer) UpdateOracleConfig(ctx context.Context, req *types.MsgUpdateOracleConfig) (*types.MsgUpdateOracleConfigResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// gov or current proposer can update metadata.
	if ms.authority != req.Authority && config.Proposer != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s or %s, got %s", ms.authority, config.Proposer, req.Authority)
	}

	config.OracleEnabled = req.OracleEnabled

	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateOracle,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOracleEnabled, strconv.FormatBool(config.OracleEnabled)),
	))
	return &types.MsgUpdateOracleConfigResponse{}, nil
}

func (ms MsgServer) UpdateMetadata(ctx context.Context, req *types.MsgUpdateMetadata) (*types.MsgUpdateMetadataResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// gov or current proposer can update metadata.
	if ms.authority != req.Authority && config.Proposer != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s or %s, got %s", ms.authority, config.Proposer, req.Authority)
	}

	config.Metadata = req.Metadata
	if err := ms.Keeper.bridgeHook.BridgeMetadataUpdated(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	finalizedOutputIndex, finalizedOutput, err := ms.GetLastFinalizedOutput(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateMetadata,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyFinalizedOutputIndex, strconv.FormatUint(finalizedOutputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyFinalizedL2BlockNumber, strconv.FormatUint(finalizedOutput.L2BlockNumber, 10)),
	))

	return &types.MsgUpdateMetadataResponse{
		OutputIndex:   finalizedOutputIndex,
		L2BlockNumber: finalizedOutput.L2BlockNumber,
	}, nil
}

// UpdateParams implements updating the parameters
func (ms MsgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	if err := ms.SetParams(ctx, *req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil

}
