package opchild

import (
	"context"

	tmtypes "github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/initia-labs/OPinit/x/opchild/keeper"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx context.Context, keeper *keeper.Keeper) (vals []tmtypes.GenesisValidator, err error) {
	err = keeper.IterateLastValidators(ctx, func(validator types.ValidatorI, power int64) (stop bool, err error) {
		pk, err := validator.ConsPubKey()
		if err != nil {
			return true, err
		}
		tmPk, err := cryptocodec.ToTmPubKeyInterface(pk)
		if err != nil {
			return true, err
		}

		vals = append(vals, tmtypes.GenesisValidator{
			Address: sdk.ConsAddress(tmPk.Address()).Bytes(),
			PubKey:  tmPk,
			Power:   validator.GetConsensusPower(),
			Name:    validator.GetMoniker(),
		})

		return false, nil
	})

	return
}
