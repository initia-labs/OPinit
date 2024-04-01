package keeper

import (
	"context"
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) RegisterExecutorChangePlan(proposalID int64, height int64, nextExecutor string, moniker string, consensusPubKey string, info string) error {
	if proposalID <= 0 {
		return errorsmod.Wrap(types.ErrInvalidExecutorChangePlan, "invalid proposal id")
	}

	if height <= 0 {
		return errorsmod.Wrap(types.ErrInvalidExecutorChangePlan, "invalid height")
	}

	if _, found := k.ExecutorChangePlans[height]; found {
		return types.ErrAlreadyRegisteredHeight
	}

	accAddr, err := k.addressCodec.StringToBytes(nextExecutor)
	if err != nil {
		return err
	}
	valAddr := sdk.ValAddress(accAddr)

	pkBytes, err := base64.StdEncoding.DecodeString(consensusPubKey)
	if err != nil {
		return err
	}

	validator, err := types.NewValidator(valAddr, &ed25519.PubKey{Key: pkBytes}, moniker)
	if err != nil {
		return err
	}

	k.ExecutorChangePlans[height] = types.ExecutorChangePlan{
		ProposalID:    proposalID,
		Height:        height,
		NextExecutor:  nextExecutor,
		NextValidator: validator,
		Info:          info,
	}
	return nil
}

func (k Keeper) ChangeExecutor(ctx context.Context, plan types.ExecutorChangePlan) error {
	err := k.Validators.Walk(ctx, nil, func(key []byte, validator types.Validator) (stop bool, err error) {
		validator.ConsPower = 0
		err = k.Validators.Set(ctx, key, validator)
		return false, err
	})
	if err != nil {
		return err
	}

	if err := k.SetValidator(ctx, plan.NextValidator); err != nil {
		return err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.BridgeExecutor = plan.NextExecutor
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}
	return nil
}
