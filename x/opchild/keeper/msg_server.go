package keeper

import (
	"bytes"
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

type MsgServer struct {
	Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServerImpl return MsgServer instance
func NewMsgServerImpl(k Keeper) MsgServer {
	return MsgServer{k}
}

// checkAdminPermission checks if the sender is the admin
func (ms MsgServer) checkAdminPermission(ctx context.Context, sender string) error {
	params, err := ms.GetParams(ctx)
	if err != nil {
		return err
	}

	if params.Admin != sender {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "the message is allowed to be executed by admin %s", params.Admin)
	}

	return nil
}

// checkBridgeExecutorPermission checks if the sender is the registered bridge executor to send messages
func (ms MsgServer) checkBridgeExecutorPermission(ctx context.Context, sender string) error {
	senderAddr, err := ms.authKeeper.AddressCodec().StringToBytes(sender)
	if err != nil {
		return err
	}

	bridgeExecutor, err := ms.BridgeExecutor(ctx)
	if err != nil {
		return err
	}

	if !bridgeExecutor.Equals(sdk.AccAddress(senderAddr)) {
		return errors.Wrapf(sdkerrors.ErrUnauthorized, "expected %s, got %s", bridgeExecutor, sender)
	}
	return nil
}

/////////////////////////////////////////////////////
// The messages for Validator

// ExecuteMessages implements a ExecuteMessages message handling
func (ms MsgServer) ExecuteMessages(ctx context.Context, req *types.MsgExecuteMessages) (*types.MsgExecuteMessagesResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	// permission check
	if err := ms.checkAdminPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, writeCache := sdkCtx.CacheContext()
	messages, err := req.GetMsgs()
	if err != nil {
		return nil, err
	}

	authority, err := ms.authKeeper.AddressCodec().StringToBytes(ms.authority)
	if err != nil {
		return nil, err
	}

	events := sdk.EmptyEvents()
	for _, msg := range messages {
		// perform a basic validation of the message
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return nil, errors.Wrap(types.ErrInvalidExecuteMsg, err.Error())
			}
		}

		signers, _, err := ms.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		if len(signers) != 1 {
			return nil, types.ErrInvalidSigner
		}

		// assert that the opchild module account is the only signer for ExecuteMessages message
		if !bytes.Equal(signers[0], authority) {
			return nil, errors.Wrapf(types.ErrInvalidSigner, sdk.AccAddress(signers[0]).String())
		}

		handler := ms.Router().Handler(msg)
		if handler == nil {
			return nil, errors.Wrap(types.ErrUnroutableExecuteMsg, sdk.MsgTypeURL(msg))
		}

		var res *sdk.Result
		res, err = handler(cacheCtx, msg)
		if err != nil {
			return nil, err
		}

		events = append(events, res.GetEvents()...)
	}

	writeCache()

	// TODO - merge events of MsgExecuteMessages itself
	sdkCtx.EventManager().EmitEvents(events)

	return &types.MsgExecuteMessagesResponse{}, nil
}

//////////////////////////////////////////////
// Authority messages

// AddValidator implements adding a validator to the designated validator set
func (ms MsgServer) AddValidator(ctx context.Context, req *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec(), ms.validatorAddressCodec); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	valAddr, err := ms.Keeper.validatorAddressCodec.StringToBytes(req.ValidatorAddress)
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

	cp := sdkCtx.ConsensusParams()
	if cp.Validator != nil {
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

	if err := ms.SetValidator(ctx, validator); err != nil {
		return nil, err
	}
	if err = ms.SetValidatorByConsAddr(ctx, validator); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, req.ValidatorAddress),
		),
	})

	return &types.MsgAddValidatorResponse{}, nil
}

// RemoveValidator implements removing a validator from the designated validator set
func (ms MsgServer) RemoveValidator(ctx context.Context, req *types.MsgRemoveValidator) (*types.MsgRemoveValidatorResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec(), ms.validatorAddressCodec); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	valAddr, err := ms.Keeper.validatorAddressCodec.StringToBytes(req.ValidatorAddress)
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
	if err := ms.Keeper.SetValidator(ctx, val); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, req.ValidatorAddress),
		),
	})

	return &types.MsgRemoveValidatorResponse{}, nil
}

// UpdateParams implements updating the parameters
func (ms MsgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// sort the min gas prices
	if req.Params != nil && req.Params.MinGasPrices != nil {
		req.Params.MinGasPrices = req.Params.MinGasPrices.Sort()
	}

	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	if err := ms.SetParams(ctx, *req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SpendFeePool implements MsgServer interface.
func (ms MsgServer) SpendFeePool(ctx context.Context, req *types.MsgSpendFeePool) (*types.MsgSpendFeePoolResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	recipientAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.Recipient)
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

func (ms MsgServer) SetBridgeInfo(ctx context.Context, req *types.MsgSetBridgeInfo) (*types.MsgSetBridgeInfoResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	// permission check
	if err := ms.checkBridgeExecutorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	// check bridge id and addr consistency
	if ok, err := ms.BridgeInfo.Has(ctx); err != nil {
		return nil, err
	} else if ok {
		info, err := ms.BridgeInfo.Get(ctx)
		if err != nil {
			return nil, err
		}

		if info.BridgeId != req.BridgeInfo.BridgeId {
			return nil, types.ErrInvalidBridgeInfo.Wrapf("expected bridge id %d, got %d", info.BridgeId, req.BridgeInfo.BridgeId)
		}

		if info.BridgeAddr != req.BridgeInfo.BridgeAddr {
			return nil, types.ErrInvalidBridgeInfo.Wrapf("expected bridge addr %s, got %s", info.BridgeAddr, req.BridgeInfo.BridgeAddr)
		}
	}

	// set bridge info
	if err := ms.BridgeInfo.Set(ctx, req.BridgeInfo); err != nil {
		return nil, err
	}

	// emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetBridgeInfo,
			sdk.NewAttribute(types.AttributeKeyBridgeId, strconv.FormatUint(req.BridgeInfo.BridgeId, 10)),
			sdk.NewAttribute(types.AttributeKeyBridgeAddr, req.BridgeInfo.BridgeAddr),
		),
	)

	return &types.MsgSetBridgeInfoResponse{}, nil
}

// FinalizeTokenDeposit implements send a deposit from the upper layer to the recipient
func (ms MsgServer) FinalizeTokenDeposit(ctx context.Context, req *types.MsgFinalizeTokenDeposit) (*types.MsgFinalizeTokenDepositResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	coin := req.Amount

	// permission check
	if err := ms.checkBridgeExecutorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	// check already finalized
	if ok, err := ms.HasFinalizedL1Sequence(ctx, req.Sequence); err != nil {
		return nil, err
	} else if ok {
		return nil, types.ErrDepositAlreadyFinalized
	}

	fromAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.From)
	if err != nil {
		return nil, err
	}

	toAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.To)
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

	if err := ms.RecordFinalizedL1Sequence(ctx, req.Sequence); err != nil {
		return nil, err
	}

	// register denom metadata
	if ok := ms.bankKeeper.HasDenomMetaData(ctx, coin.Denom); !ok {
		ms.setDenomMetadata(ctx, req.BaseDenom, coin.Denom)
	}

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
		subCtx, commit := sdkCtx.CacheContext()

		err = ms.bridgeHook(subCtx, fromAddr, req.Data)
		if err == nil {
			commit()
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookSuccess, "true"))
		} else {
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookSuccess, "false"))
		}
	}
	sdkCtx.EventManager().EmitEvent(event)

	return &types.MsgFinalizeTokenDepositResponse{}, nil
}

/////////////////////////////////////////////////////
// The messages for User

// InitiateTokenWithdrawal implements creating a token from the upper layer
func (ms MsgServer) InitiateTokenWithdrawal(ctx context.Context, req *types.MsgInitiateTokenWithdrawal) (*types.MsgInitiateTokenWithdrawalResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	coin := req.Amount

	senderAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.Sender)
	if err != nil {
		return nil, err
	}

	coins := sdk.NewCoins(coin)
	l2Sequence, err := ms.IncreaseNextL2Sequence(ctx)
	if err != nil {
		return nil, err
	}
	if err := ms.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := ms.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, req.Sender),
		sdk.NewAttribute(types.AttributeKeyTo, req.To),
		sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, strconv.FormatUint(l2Sequence, 10)),
	))

	return &types.MsgInitiateTokenWithdrawalResponse{}, nil
}

func (ms MsgServer) UpdateOracle(ctx context.Context, req *types.MsgUpdateOracle) (*types.MsgUpdateOracleResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	// permission check
	if err := ms.checkBridgeExecutorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	err := ms.Keeper.ApplyOracleUpdate(ctx, req.Height, req.Data)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateOracle,
		sdk.NewAttribute(types.AttributeKeyFrom, req.Sender),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatUint(req.Height, 10)),
	))

	return &types.MsgUpdateOracleResponse{}, nil
}
