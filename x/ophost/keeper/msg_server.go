package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"golang.org/x/crypto/sha3"

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
func (ms MsgServer) RecordBatch(context context.Context, req *types.MsgRecordBatch) (*types.MsgRecordBatchResponse, error) {
	sdk.UnwrapSDKContext(context).EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecordBatch,
			sdk.NewAttribute(types.AttributeKeySubmitter, req.Submitter),
		),
	)

	return &types.MsgRecordBatchResponse{}, nil
}

/////////////////////////////////////////////////////
// The messages for Bridge Creator

func (ms MsgServer) CreateBridge(context context.Context, req *types.MsgCreateBridge) (*types.MsgCreateBridgeResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	bridgeId := ms.IncreaseNextBridgeId(ctx)
	err := ms.SetBridgeConfig(ctx, bridgeId, req.Config)
	if err != nil {
		return nil, err
	}

	// create bridge account
	ms.authKeeper.SetAccount(ctx, types.NewBridgeAccountWithAddress(types.BridgeAddress(bridgeId)))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreateBridge,
		sdk.NewAttribute(types.AttributeKeyCreator, req.Creator),
		sdk.NewAttribute(types.AttributeKeyProposer, req.Config.Proposer),
		sdk.NewAttribute(types.AttributeKeyChallenger, req.Config.Challenger),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
	))

	// TODO in initia app
	// permit to create bridge only when the ibc channel has
	// GetNextL1SequenceSend == 1
	if err := ms.bridgeHook.BridgeCreated(ctx, bridgeId, req.Config); err != nil {
		return nil, err
	}

	return &types.MsgCreateBridgeResponse{
		BridgeId: bridgeId,
	}, nil
}

func (ms MsgServer) ProposeOutput(context context.Context, req *types.MsgProposeOutput) (*types.MsgProposeOutputResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

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
	outputIndex := ms.IncreaseNextOutputIndex(ctx, bridgeId)

	// store output proposal
	if err := ms.SetOutputProposal(ctx, bridgeId, outputIndex, types.Output{
		OutputRoot:    outputRoot,
		L1BlockTime:   ctx.BlockTime(),
		L2BlockNumber: l2BlockNumber,
	}); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeProposeOutput,
		sdk.NewAttribute(types.AttributeKeyProposer, proposer),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyL2BlockNumber, strconv.FormatUint(l2BlockNumber, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputRoot, hex.EncodeToString(outputRoot)),
	))

	return &types.MsgProposeOutputResponse{
		OutputIndex: outputIndex,
	}, nil
}

func (ms MsgServer) DeleteOutput(context context.Context, req *types.MsgDeleteOutput) (*types.MsgDeleteOutputResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	challenger := req.Challenger
	bridgeId := req.BridgeId
	outputIndex := req.OutputIndex

	bridgeConfig, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	// permission check
	if challenger != bridgeConfig.Challenger {
		return nil, errors.ErrUnauthorized.Wrap("invalid challenger")
	}

	// delete output proposal
	ms.DeleteOutputProposal(ctx, bridgeId, outputIndex)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeleteOutput,
		sdk.NewAttribute(types.AttributeKeyChallenger, challenger),
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
	))

	return &types.MsgDeleteOutputResponse{}, nil
}

func (ms MsgServer) InitiateTokenDeposit(context context.Context, req *types.MsgInitiateTokenDeposit) (*types.MsgInitiateTokenDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	coin := req.Amount
	bridgeId := req.BridgeId
	l1Sequence := ms.IncreaseNextL1Sequence(ctx, bridgeId)

	// send the funds to bridge address
	bridgeAddr := types.BridgeAddress(bridgeId)
	if err := ms.bankKeeper.SendCoins(ctx, sender, bridgeAddr, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}

	// record token pairs
	l2Denom := types.L2Denom(bridgeId, coin.Denom)
	if !ms.HasTokenPair(ctx, bridgeId, l2Denom) {
		ms.SetTokenPair(ctx, bridgeId, l2Denom, coin.Denom)
	}

	// emit events for bridge executor
	ctx.EventManager().EmitEvent(sdk.NewEvent(
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

	return &types.MsgInitiateTokenDepositResponse{}, nil
}

func (ms MsgServer) FinalizeTokenWithdrawal(context context.Context, req *types.MsgFinalizeTokenWithdrawal) (*types.MsgFinalizeTokenWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}
	receiver, err := sdk.AccAddressFromBech32(req.Receiver)
	if err != nil {
		return nil, err
	}

	bridgeId := req.BridgeId
	outputIndex := req.OutputIndex
	l2Sequence := req.Sequence
	amount := req.Amount.Amount
	l2Denom := req.Amount.Denom

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
	{
		seed := make([]byte, 32*4)
		copy(seed, req.Version)
		copy(seed[32:], req.StateRoot)
		copy(seed[64:], req.StorageRoot)
		copy(seed[96:], req.LatestBlockHash)
		outputRoot := sha3.Sum256(seed)

		if !bytes.Equal(outputProposal.OutputRoot, outputRoot[:]) {
			return nil, types.ErrFailedToVerifyWithdrawal.Wrap("invalid output root")
		}
	}

	// verify storage root can be generated from
	// withdrawal proofs and withdrawal tx data.
	{
		var withdrawalHash [32]byte
		{
			seed := []byte{}
			binary.BigEndian.AppendUint64(seed, bridgeId)
			binary.BigEndian.AppendUint64(seed, req.Sequence)
			seed = append(seed, sender[:]...)
			seed = append(seed, receiver[:]...)
			seed = append(seed, []byte(l2Denom)...)
			binary.BigEndian.AppendUint64(seed, amount.Uint64())

			withdrawalHash = sha3.Sum256(seed)
		}

		if ms.HasProvenWithdrawal(ctx, bridgeId, withdrawalHash) {
			return nil, types.ErrWithdrawalAlreadyFinalized
		}

		// should works with sorted merkle tree
		rootSeed := withdrawalHash
		proofs := req.WithdrawalProofs
		for _, proof := range proofs {
			switch bytes.Compare(rootSeed[:], proof) {
			case 0, 1: // equal or greater
				rootSeed = sha3.Sum256(append(proof, rootSeed[:]...))
			case -1: // less
				rootSeed = sha3.Sum256(append(rootSeed[:], proof...))
			}
		}

		rootHash := rootSeed
		if !bytes.Equal(req.StorageRoot, rootHash[:]) {
			return nil, types.ErrFailedToVerifyWithdrawal.Wrap("invalid storage root proofs")
		}

		ms.RecordProvenWithdrawal(ctx, bridgeId, withdrawalHash)
	}

	// load l1denom from the token pair store
	l1Denom, err := ms.GetTokenPair(ctx, bridgeId, l2Denom)
	if err != nil {
		return nil, err
	}

	// transfer asset to a user from the bridge account
	bridgeAddr := types.BridgeAddress(bridgeId)
	if err := ms.bankKeeper.SendCoins(ctx, bridgeAddr, receiver, sdk.NewCoins(sdk.NewCoin(l1Denom, amount))); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeFinalizeTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(bridgeId, 10)),
		sdk.NewAttribute(types.AttributeKeyOutputIndex, strconv.FormatUint(outputIndex, 10)),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, strconv.FormatUint(l2Sequence, 10)),
		sdk.NewAttribute(types.AttributeKeyFrom, sender.String()),
		sdk.NewAttribute(types.AttributeKeyTo, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyL1Denom, l1Denom),
		sdk.NewAttribute(types.AttributeKeyL2Denom, l2Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))

	return &types.MsgFinalizeTokenWithdrawalResponse{}, nil
}

func (ms MsgServer) UpdateProposer(context context.Context, req *types.MsgUpdateProposer) (*types.MsgUpdateProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	config.Proposer = req.NewProposer
	ms.Keeper.bridgeHook.BridgeProposerUpdated(ctx, bridgeId, config)
	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	return &types.MsgUpdateProposerResponse{}, nil
}

func (ms MsgServer) UpdateChallenger(context context.Context, req *types.MsgUpdateChallenger) (*types.MsgUpdateChallengerResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	bridgeId := req.BridgeId
	config, err := ms.GetBridgeConfig(ctx, bridgeId)
	if err != nil {
		return nil, err
	}

	config.Challenger = req.NewChallenger
	ms.Keeper.bridgeHook.BridgeChallengerUpdated(ctx, bridgeId, config)
	if err := ms.SetBridgeConfig(ctx, bridgeId, config); err != nil {
		return nil, err
	}

	return &types.MsgUpdateChallengerResponse{}, nil
}

// UpdateParams implements updating the parameters
func (ms MsgServer) UpdateParams(context context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(context)
	if err := ms.SetParams(ctx, *req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil

}
