package keeper

import (
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
	bridgeHook types.BridgeHook

	// Msg server router
	router *baseapp.MsgServiceRouter

	// Legacy Proposal router
	legacyRouter govv1beta1.Router

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/opchild module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	bh types.BridgeHook,
	router *baseapp.MsgServiceRouter,
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
		bridgeHook: bh,
		router:     router,
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

// Router returns the gov keeper's router
func (keeper Keeper) Router() *baseapp.MsgServiceRouter {
	return keeper.router
}

// SetLegacyRouter sets the legacy router for governance(validator operation)
func (keeper *Keeper) SetLegacyRouter(router govv1beta1.Router) {
	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	router.Seal()
	keeper.legacyRouter = router
}

// LegacyRouter returns the rollup keeper's legacy router
func (keeper Keeper) LegacyRouter() govv1beta1.Router {
	return keeper.legacyRouter
}
