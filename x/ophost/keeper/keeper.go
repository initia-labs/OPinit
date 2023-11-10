package keeper

import (
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
	bridgeHook types.BridgeHook

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	bridgeHook types.BridgeHook,
	authority string,
) Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   key,
		authKeeper: ak,
		bankKeeper: bk,
		bridgeHook: bridgeHook,
		authority:  authority,
	}
}

// GetAuthority returns the x/move module's authority.
func (ak Keeper) GetAuthority() string {
	return ak.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
