package keeper

import (
	"context"
	"strconv"

	"cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

const (
	bridgeModuleName           = "op_bridge"
	bridgeFinalizeFunctionName = "finalize_token_bridge"
	bridgeRegisterFunctionName = "register_token"
)

type MsgServer struct {
	Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServerImpl return MsgServer instance
func NewMsgServerImpl(k Keeper) MsgServer {
	return MsgServer{k}
}

// checkValidatorPermission checks if the sender is the one of validator
func (ms MsgServer) checkValidatorPermission(ctx sdk.Context, sender string) error {
	addr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	valAddr := sdk.ValAddress(addr)
	if _, found := ms.GetValidator(ctx, valAddr); !found {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "the message is allowed to be executed by validator")
	}

	return nil
}

// checkBridgeExecutorPermission checks if the sender is the registered bridge executor to send messages
func (ms MsgServer) checkBridgeExecutorPermission(ctx sdk.Context, sender string) error {
	senderAddr, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	bridgeExecutor := ms.BridgeExecutor(ctx)
	if !bridgeExecutor.Equals(senderAddr) {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "expected %s, got %s", bridgeExecutor, sender)
	}
	return nil
}

/////////////////////////////////////////////////////
// The messages for Validator

// ExecuteMessages implements a ExecuteMessages message handling
func (ms MsgServer) ExecuteMessages(context context.Context, req *types.MsgExecuteMessages) (*types.MsgExecuteMessagesResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	authority := sdk.MustAccAddressFromBech32(ms.authority)

	// permission check
	if err := ms.checkValidatorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	cachtCtx, writeCache := ctx.CacheContext()
	messages, err := req.GetMsgs()
	if err != nil {
		return nil, err
	}

	events := sdk.EmptyEvents()
	for _, msg := range messages {
		signers := msg.GetSigners()
		// assert that the rollup module account is the only signer for ExecuteMessages message
		if !signers[0].Equals(authority) {
			return nil, errors.Wrapf(types.ErrInvalidSigner, signers[0].String())
		}

		handler := ms.Router().Handler(msg)

		var res *sdk.Result
		res, err = handler(cachtCtx, msg)
		if err != nil {
			break
		}

		events = append(events, res.GetEvents()...)

	}
	if err != nil {
		return nil, err
	}
	writeCache()

	// TODO - merge events of MsgExecuteMessages itself
	ctx.EventManager().EmitEvents(events)

	return &types.MsgExecuteMessagesResponse{}, nil
}

// ExecuteLegacyContents implements a ExecuteLegacyContents message handling
func (ms MsgServer) ExecuteLegacyContents(context context.Context, req *types.MsgExecuteLegacyContents) (*types.MsgExecuteLegacyContentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	for _, content := range req.GetContents() {
		// Ensure that the content has a respective handler
		if !ms.Keeper.legacyRouter.HasRoute(content.ProposalRoute()) {
			return nil, errors.Wrap(govtypes.ErrNoProposalHandlerExists, content.ProposalRoute())
		}

		handler := ms.Keeper.legacyRouter.GetRoute(content.ProposalRoute())
		if err := handler(ctx, content); err != nil {
			return nil, errors.Wrapf(govtypes.ErrInvalidProposalContent, "failed to run legacy handler %s, %+v", content.ProposalRoute(), err)
		}
	}

	return &types.MsgExecuteLegacyContentsResponse{}, nil
}

//////////////////////////////////////////////
// Authority messages

// AddValidator implements adding a validator to the designated validator set
func (ms MsgServer) AddValidator(context context.Context, req *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(context)
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// check to see if the pubkey or sender has been registered before
	if _, found := ms.GetValidator(ctx, valAddr); found {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := req.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, found := ms.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); found {
		return nil, types.ErrValidatorPubKeyExists
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		pkType := pk.Type()
		hasKeyType := false
		for _, keyType := range cp.Validator.PubKeyTypes {
			if pkType == keyType {
				hasKeyType = true
				break
			}
		}
		if !hasKeyType {
			return nil, errors.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(valAddr, pk, req.Moniker)
	if err != nil {
		return nil, err
	}

	ms.SetValidator(ctx, validator)
	if err = ms.SetValidatorByConsAddr(ctx, validator); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, req.ValidatorAddress),
		),
	})

	return &types.MsgAddValidatorResponse{}, nil
}

// RemoveValidator implements removing a validator from the designated validator set
func (ms MsgServer) RemoveValidator(context context.Context, req *types.MsgRemoveValidator) (*types.MsgRemoveValidatorResponse, error) {
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(context)
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	val, found := ms.Keeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, errors.Wrap(types.ErrNoValidatorFound, val.OperatorAddress)
	}
	val.ConsPower = 0

	// set validator consensus power `0`,
	// then `val_state_change` will execute `k.RemoveValidator`.
	ms.Keeper.SetValidator(ctx, val)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, req.ValidatorAddress),
		),
	})

	return &types.MsgRemoveValidatorResponse{}, nil
}

// UpdateParams implements updating the parameters
func (ms MsgServer) UpdateParams(context context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(context)
	if err := ms.SetParams(ctx, *req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil

}

// SpendFeePool implements MsgServer interface.
func (ms MsgServer) SpendFeePool(context context.Context, req *types.MsgSpendFeePool) (*types.MsgSpendFeePoolResponse, error) {
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(context)
	recipientAddr, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		return nil, err
	}

	// send collected fees to the recipient address
	if err := ms.bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, recipientAddr, req.Amount); err != nil {
		return nil, err
	}

	return &types.MsgSpendFeePoolResponse{}, nil
}

/////////////////////////////////////////////////////
// The messages for Bridge Executor

// FinalizeTokenDeposit implements send a deposit from the upper layer to the recipient
func (ms MsgServer) FinalizeTokenDeposit(context context.Context, req *types.MsgFinalizeTokenDeposit) (*types.MsgFinalizeTokenDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	coin := req.Amount

	// permission check
	if err := ms.checkBridgeExecutorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	// check already finalized
	if ok := ms.HasFinalizedL1Sequence(ctx, req.Sequence); ok {
		return nil, types.ErrDepositAlreadyFinalized
	}

	fromAddr, err := sdk.AccAddressFromBech32(req.From)
	if err != nil {
		return nil, err
	}

	toAddr, err := sdk.AccAddressFromBech32(req.To)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(coin)
	if err := ms.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	if err := ms.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, coins); err != nil {
		return nil, err
	}

	ms.RecordFinalizedL1Sequence(ctx, req.Sequence)

	event := sdk.NewEvent(
		types.EventTypeFinalizeTokenDeposit,
		sdk.NewAttribute(types.AttributeKeyL1Sequence, strconv.FormatUint(req.Sequence, 10)),
		sdk.NewAttribute(types.AttributeKeySender, req.From),
		sdk.NewAttribute(types.AttributeKeyRecipient, req.To),
		sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyFinalizeHeight, strconv.FormatUint(req.Height, 10)),
	)

	// handle hook
	if len(req.Data) > 0 {
		subCtx, commit := ctx.CacheContext()

		err = ms.bridgeHook(subCtx, fromAddr, req.Data)
		if err == nil {
			commit()
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookSuccess, "true"))
		} else {
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookSuccess, "false"))
		}
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgFinalizeTokenDepositResponse{}, nil
}

/////////////////////////////////////////////////////
// The messages for User

// InitiateTokenWithdrawal implements creating a token from the upper layer
func (ms MsgServer) InitiateTokenWithdrawal(context context.Context, req *types.MsgInitiateTokenWithdrawal) (*types.MsgInitiateTokenWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)
	coin := req.Amount

	senderAddr, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(coin)
	l2Sequence := ms.IncreaseNextL2Sequence(ctx)
	if err := ms.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := ms.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, req.Sender),
		sdk.NewAttribute(types.AttributeKeyTo, req.To),
		sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, strconv.FormatUint(l2Sequence, 10)),
	))

	return &types.MsgInitiateTokenWithdrawalResponse{}, nil
}
