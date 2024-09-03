package keeper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

type MsgServer struct {
	*Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServerImpl return MsgServer instance
func NewMsgServerImpl(k *Keeper) *MsgServer {
	return &MsgServer{k}
}

// checkAdminPermission checks if the sender is the admin
func (ms MsgServer) checkAdminPermission(ctx context.Context, sender string) error {
	params, err := ms.GetParams(ctx)
	if err != nil {
		return err
	}

	if params.Admin != sender {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "the message is allowed to be executed by admin %s", params.Admin)
	}

	return nil
}

// checkBridgeExecutorPermission checks if the sender is the registered bridge executor to send messages
func (ms MsgServer) checkBridgeExecutorPermission(ctx context.Context, sender string) error {
	senderAddr, err := ms.authKeeper.AddressCodec().StringToBytes(sender)
	if err != nil {
		return err
	}

	bridgeExecutors, err := ms.BridgeExecutors(ctx)
	if err != nil {
		return err
	}
	isIncluded := false
	for _, bridgeExecutor := range bridgeExecutors {
		if bytes.Equal(bridgeExecutor, senderAddr) {
			isIncluded = true
		}
	}
	if !isIncluded {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "expected included in %s, got %s", bridgeExecutors, sender)
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
				return nil, errorsmod.Wrap(types.ErrInvalidExecuteMsg, err.Error())
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
			signer, err := ms.addressCodec.BytesToString(signers[0])
			if err != nil {
				return nil, err
			}

			return nil, errorsmod.Wrap(types.ErrInvalidSigner, signer)
		}

		handler := ms.Router().Handler(msg)
		if handler == nil {
			return nil, errorsmod.Wrap(types.ErrUnroutableExecuteMsg, sdk.MsgTypeURL(msg))
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
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	allValidators, err := ms.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	numMaxValidators, err := ms.MaxValidators(ctx)
	if err != nil {
		return nil, err
	}

	if int(numMaxValidators) <= len(allValidators) {
		return nil, types.ErrMaxValidatorsExceeded
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
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
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
			return nil, errorsmod.Wrapf(
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
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	valAddr, err := ms.Keeper.validatorAddressCodec.StringToBytes(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	val, found := ms.Keeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, errorsmod.Wrap(types.ErrNoValidatorFound, val.OperatorAddress)
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
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
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
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
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

		if info.L1ChainId != req.BridgeInfo.L1ChainId {
			return nil, types.ErrInvalidBridgeInfo.Wrapf("expected l1 chain id %s, got %s", info.L1ChainId, req.BridgeInfo.L1ChainId)
		}

		if info.L1ClientId != "" && info.L1ClientId != req.BridgeInfo.L1ClientId {
			return nil, types.ErrInvalidBridgeInfo.Wrapf("expected l1 client id %s, got %s", info.L1ClientId, req.BridgeInfo.L1ClientId)
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
			sdk.NewAttribute(types.AttributeKeyL1ChainId, req.BridgeInfo.L1ChainId),
			sdk.NewAttribute(types.AttributeKeyL1ClientId, req.BridgeInfo.L1ClientId),
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

	finalizedL1Sequence, err := ms.GetNextL1Sequence(ctx)
	if err != nil {
		return nil, err
	}

	if req.Sequence < finalizedL1Sequence {
		// No op instead of returning an error
		return &types.MsgFinalizeTokenDepositResponse{Result: types.NOOP}, nil
	} else if req.Sequence > finalizedL1Sequence {
		return nil, types.ErrInvalidSequence
	}

	// deposit token
	var success bool
	var reason string
	toAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.To)
	if err != nil {
		success = false
		reason = fmt.Sprintf("failed to convert recipient address: %s", err)
	} else {
		// rollback if the deposit is failed
		success, reason = ms.safeDepositToken(ctx, toAddr, sdk.NewCoins(coin))
	}

	// updae l1 sequence
	if _, err := ms.IncreaseNextL1Sequence(ctx); err != nil {
		return nil, err
	}

	// register denom metadata
	if ok := ms.bankKeeper.HasDenomMetaData(ctx, coin.Denom); !ok {
		ms.setDenomMetadata(ctx, req.BaseDenom, coin.Denom)
	}

	// register denom pair
	if ok, err := ms.DenomPairs.Has(ctx, coin.Denom); err != nil {
		return nil, err
	} else if !ok {
		if err := ms.DenomPairs.Set(ctx, coin.Denom, req.BaseDenom); err != nil {
			return nil, err
		}
	}

	event := sdk.NewEvent(
		types.EventTypeFinalizeTokenDeposit,
		sdk.NewAttribute(types.AttributeKeyL1Sequence, strconv.FormatUint(req.Sequence, 10)),
		sdk.NewAttribute(types.AttributeKeySender, req.From),
		sdk.NewAttribute(types.AttributeKeyRecipient, req.To),
		sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, req.BaseDenom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyFinalizeHeight, strconv.FormatUint(req.Height, 10)),
		sdk.NewAttribute(types.AttributeKeySuccess, strconv.FormatBool(success)),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	)

	// if the deposit is successful and the data is not empty, execute the hook
	if success && len(req.Data) > 0 {
		success, reason := ms.handleBridgeHook(sdkCtx, req.Data)
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookSuccess, strconv.FormatBool(success)))
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyHookReason, reason))
	}

	// emit deposit event
	sdkCtx.EventManager().EmitEvent(event)

	// if the deposit is failed, initate a withdrawal
	if !success {
		l2Sequence, err := ms.IncreaseNextL2Sequence(ctx)
		if err != nil {
			return nil, err
		}

		err = ms.emitWithdrawEvents(ctx, types.NewMsgInitiateTokenWithdrawal(req.To, req.From, coin), l2Sequence)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgFinalizeTokenDepositResponse{Result: types.SUCCESS}, nil
}

/////////////////////////////////////////////////////
// The messages for User

// InitiateTokenWithdrawal implements creating a token from the upper layer
func (ms MsgServer) InitiateTokenWithdrawal(ctx context.Context, req *types.MsgInitiateTokenWithdrawal) (*types.MsgInitiateTokenWithdrawalResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	coin := req.Amount
	burnCoins := sdk.NewCoins(coin)

	senderAddr, err := ms.authKeeper.AddressCodec().StringToBytes(req.Sender)
	if err != nil {
		return nil, err
	}

	// send coins to the module account only if the amount is positive
	// - pending deposits are already accounted for
	if coin.IsPositive() {
		if err := ms.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(coin)); err != nil {
			return nil, err
		}
	}

	// burn withdrawn coins from the module account
	if err := ms.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
		return nil, err
	}

	l2Sequence, err := ms.IncreaseNextL2Sequence(ctx)
	if err != nil {
		return nil, err
	}

	err = ms.emitWithdrawEvents(ctx, req, l2Sequence)
	if err != nil {
		return nil, err
	}

	return &types.MsgInitiateTokenWithdrawalResponse{
		Sequence: l2Sequence,
	}, nil
}

func (ms MsgServer) emitWithdrawEvents(ctx context.Context, req *types.MsgInitiateTokenWithdrawal, l2Sequence uint64) error {
	coin := req.Amount
	baseDenom, err := ms.GetBaseDenom(ctx, coin.Denom)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInitiateTokenWithdrawal,
		sdk.NewAttribute(types.AttributeKeyFrom, req.Sender),
		sdk.NewAttribute(types.AttributeKeyTo, req.To),
		sdk.NewAttribute(types.AttributeKeyDenom, coin.Denom),
		sdk.NewAttribute(types.AttributeKeyBaseDenom, baseDenom),
		sdk.NewAttribute(types.AttributeKeyAmount, coin.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyL2Sequence, strconv.FormatUint(l2Sequence, 10)),
	))

	return nil
}

func (ms MsgServer) UpdateOracle(ctx context.Context, req *types.MsgUpdateOracle) (*types.MsgUpdateOracleResponse, error) {
	if err := req.Validate(ms.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	// permission check
	if err := ms.checkBridgeExecutorPermission(ctx, req.Sender); err != nil {
		return nil, err
	}

	// config check
	info, err := ms.Keeper.BridgeInfo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !info.BridgeConfig.OracleEnabled {
		return nil, types.ErrOracleDisabled
	}

	err = ms.Keeper.ApplyOracleUpdate(ctx, req.Height, req.Data)
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
