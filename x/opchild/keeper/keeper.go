package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/initia-labs/OPinit/x/opchild/types"
)

var _ types.AnteKeeper = Keeper{}

type Keeper struct {
	cdc          codec.Codec
	storeService corestoretypes.KVStoreService

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
	bridgeHook types.BridgeHook

	// Msg server router
	router *baseapp.MsgServiceRouter

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/opchild module account.
	authority string

	addressCodec          address.Codec
	validatorAddressCodec address.Codec
	consensusAddressCodec address.Codec

	Schema               collections.Schema
	NextL2Sequence       collections.Sequence
	Params               collections.Item[types.Params]
	FinalizedL1Sequence  collections.Map[uint64, bool]
	LastValidatorPowers  collections.Map[[]byte, int64]
	Validators           collections.Map[[]byte, types.Validator]
	ValidatorsByConsAddr collections.Map[[]byte, []byte]
	HistoricalInfos      collections.Map[int64, cosmostypes.HistoricalInfo]

	ExecutorChangePlans map[uint64]types.ExecutorChangePlan
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	bh types.BridgeHook,
	router *baseapp.MsgServiceRouter,
	authority string,
	addressCodec address.Codec,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
) *Keeper {

	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := &Keeper{
		cdc:                   cdc,
		storeService:          storeService,
		authKeeper:            ak,
		bankKeeper:            bk,
		bridgeHook:            bh,
		router:                router,
		authority:             authority,
		addressCodec:          addressCodec,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		NextL2Sequence:        collections.NewSequence(sb, types.NextL2SequenceKey, "next_l2_sequence"),
		Params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		FinalizedL1Sequence:   collections.NewMap(sb, types.FinalizedL1SequencePrefix, "finalized_l1_sequence", collections.Uint64Key, collections.BoolValue),
		LastValidatorPowers:   collections.NewMap(sb, types.LastValidatorPowerPrefix, "last_validator_powers", collections.BytesKey, collections.Int64Value),
		Validators:            collections.NewMap(sb, types.ValidatorsPrefix, "validators", collections.BytesKey, codec.CollValue[types.Validator](cdc)),
		ValidatorsByConsAddr:  collections.NewMap(sb, types.ValidatorsByConsAddrPrefix, "validators_by_cons_addr", collections.BytesKey, collections.BytesValue),
		HistoricalInfos:       collections.NewMap(sb, types.HistoricalInfoPrefix, "historical_infos", collections.Int64Key, codec.CollValue[cosmostypes.HistoricalInfo](cdc)),

		ExecutorChangePlans: make(map[uint64]types.ExecutorChangePlan),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the x/move module's authority.
func (ak Keeper) GetAuthority() string {
	return ak.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov keeper's router
func (keeper Keeper) Router() *baseapp.MsgServiceRouter {
	return keeper.router
}
