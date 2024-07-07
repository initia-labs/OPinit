package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"

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
	NextL1Sequence       collections.Sequence
	NextL2Sequence       collections.Sequence
	Params               collections.Item[types.Params]
	BridgeInfo           collections.Item[types.BridgeInfo]
	LastValidatorPowers  collections.Map[[]byte, int64]
	Validators           collections.Map[[]byte, types.Validator]
	ValidatorsByConsAddr collections.Map[[]byte, []byte]
	HistoricalInfos      collections.Map[int64, cosmostypes.HistoricalInfo]
	DenomPairs           collections.Map[string, string]

	ExecutorChangePlans map[uint64]types.ExecutorChangePlan

	l2OracleHandler    *L2OracleHandler
	HostValidatorStore *HostValidatorStore
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	bh types.BridgeHook,
	ok *oraclekeeper.Keeper,
	router *baseapp.MsgServiceRouter,
	authority string,
	addressCodec address.Codec,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
	logger log.Logger,
) *Keeper {
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	sb := collections.NewSchemaBuilder(storeService)

	hostValidatorStore := NewHostValidatorStore(
		collections.NewItem(sb, types.HostHeightKey, "host_height", collections.Int64Value),
		collections.NewMap(sb, types.HostValidatorsPrefix, "host_validators", collections.BytesKey, codec.CollValue[cosmostypes.Validator](cdc)),
		consensusAddressCodec,
	)

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
		NextL1Sequence:        collections.NewSequence(sb, types.NextL1SequenceKey, "finalized_l1_sequence"),
		NextL2Sequence:        collections.NewSequence(sb, types.NextL2SequenceKey, "next_l2_sequence"),
		Params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		BridgeInfo:            collections.NewItem(sb, types.BridgeInfoKey, "bridge_info", codec.CollValue[types.BridgeInfo](cdc)),
		LastValidatorPowers:   collections.NewMap(sb, types.LastValidatorPowerPrefix, "last_validator_powers", collections.BytesKey, collections.Int64Value),
		Validators:            collections.NewMap(sb, types.ValidatorsPrefix, "validators", collections.BytesKey, codec.CollValue[types.Validator](cdc)),
		ValidatorsByConsAddr:  collections.NewMap(sb, types.ValidatorsByConsAddrPrefix, "validators_by_cons_addr", collections.BytesKey, collections.BytesValue),
		HistoricalInfos:       collections.NewMap(sb, types.HistoricalInfoPrefix, "historical_infos", collections.Int64Key, codec.CollValue[cosmostypes.HistoricalInfo](cdc)),
		DenomPairs:            collections.NewMap(sb, types.DenomPairPrefix, "denom_pairs", collections.StringKey, collections.StringValue),

		ExecutorChangePlans: make(map[uint64]types.ExecutorChangePlan),
		HostValidatorStore:  hostValidatorStore,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	k.l2OracleHandler = NewL2OracleHandler(k, ok, logger)

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

// setDenomMetadata sets an OPinit token's denomination metadata
func (k Keeper) setDenomMetadata(ctx context.Context, baseDenom, denom string) {
	metadata := banktypes.Metadata{
		Base:        denom,
		Display:     baseDenom,
		Symbol:      baseDenom,
		Name:        fmt.Sprintf("%s OPinit token", baseDenom),
		Description: fmt.Sprintf("OPinit token of %s", baseDenom),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    baseDenom,
				Exponent: 0,
			},
		},
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)
}

// UpdateHostValidatorSet updates the host validator set.
func (k Keeper) UpdateHostValidatorSet(ctx context.Context, clientID string, height int64, validatorSet *cmtproto.ValidatorSet) error {
	if clientID == "" {
		return nil
	}

	// ignore if the chain ID is not the host chain ID
	if l1ClientId, err := k.L1ClientId(ctx); err != nil {
		return err
	} else if l1ClientId != clientID {
		return nil
	}

	return k.HostValidatorStore.UpdateValidators(ctx, height, validatorSet)
}

// ApplyOracleUpdate applies an oracle update to the L2 oracle handler.
func (k Keeper) ApplyOracleUpdate(ctx context.Context, height uint64, extCommitBz []byte) error {
	return k.l2OracleHandler.UpdateOracle(ctx, height, extCommitBz)
}

func (k Keeper) L1ClientId(ctx context.Context) (string, error) {
	info, err := k.BridgeInfo.Get(ctx)
	if err != nil {
		return "", err
	}

	return info.L1ClientId, nil
}

func (k Keeper) L1ChainId(ctx context.Context) (string, error) {
	info, err := k.BridgeInfo.Get(ctx)
	if err != nil {
		return "", err
	}

	return info.L1ChainId, nil
}
