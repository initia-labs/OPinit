package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

func (k Keeper) RegisterExecutorChangePlan(
	proposalID, height uint64,
	nextValidator, moniker, consensusPubKey, info string, nextExecutors []string,
) error {
	if proposalID <= 0 {
		return errorsmod.Wrap(types.ErrInvalidExecutorChangePlan, "invalid proposal id")
	}

	if height <= 0 {
		return errorsmod.Wrap(types.ErrInvalidExecutorChangePlan, "invalid height")
	}

	if _, found := k.ExecutorChangePlans[height]; found {
		return types.ErrAlreadyRegisteredHeight
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(nextValidator)
	if err != nil {
		return err
	}
	for _, nextExecutor := range nextExecutors {
		_, err = k.addressCodec.StringToBytes(nextExecutor)
		if err != nil {
			return err
		}
	}

	var pubKey cryptotypes.PubKey
	err = k.cdc.UnmarshalInterfaceJSON([]byte(consensusPubKey), &pubKey)
	if err != nil {
		return errorsmod.Wrap(types.ErrInvalidExecutorChangePlan, "invalid pub key")
	}

	validator, err := types.NewValidator(valAddr, pubKey, moniker)
	if err != nil {
		return err
	}

	k.ExecutorChangePlans[height] = types.ExecutorChangePlan{
		ProposalID:    proposalID,
		Height:        height,
		NextExecutors: nextExecutors,
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
	if err = k.SetValidatorByConsAddr(ctx, plan.NextValidator); err != nil {
		return err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.BridgeExecutors = plan.NextExecutors
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}
	return nil
}
