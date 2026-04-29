package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeService corestoretypes.KVStoreService
	msgRouter    baseapp.MessageRouter

	authKeeper          types.AccountKeeper
	bankKeeper          types.BankKeeper
	bridgeHook          types.BridgeHook
	communityPoolKeeper types.CommunityPoolKeeper
	channelKeeper       types.ChannelKeeper
	oracleKeeper        types.OracleKeeper
	transferKeeper      types.TransferKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	validatorAddressCodec address.Codec

	Schema            collections.Schema
	NextBridgeId      collections.Sequence
	Params            collections.Item[types.Params]
	BridgeConfigs     collections.Map[uint64, types.BridgeConfig]
	BatchInfos        collections.Map[collections.Pair[uint64, uint64], types.BatchInfoWithOutput]
	NextL1Sequences   collections.Map[uint64, uint64]
	TokenPairs        collections.Map[collections.Pair[uint64, string], string]
	OutputProposals   collections.Map[collections.Pair[uint64, uint64], types.Output]
	NextOutputIndexes collections.Map[uint64, uint64]
	ProvenWithdrawals collections.Map[collections.Pair[uint64, []byte], bool]
	MigrationInfos    collections.Map[collections.Pair[uint64, string], types.MigrationInfo]
	OraclePriceHash   collections.Item[types.OraclePriceHash]
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	msgRouter baseapp.MessageRouter,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.CommunityPoolKeeper,
	channelKeeper types.ChannelKeeper,
	oracleKeeper types.OracleKeeper,
	transferKeeper types.TransferKeeper,
	bridgeHook types.BridgeHook,
	authority string,
	validatorAddressCodec address.Codec,
) *Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := &Keeper{
		cdc:          cdc,
		storeService: storeService,
		msgRouter:    msgRouter,

		authKeeper:          ak,
		bankKeeper:          bk,
		communityPoolKeeper: ck,
		channelKeeper:       channelKeeper,
		oracleKeeper:        oracleKeeper,
		transferKeeper:      transferKeeper,

		bridgeHook: bridgeHook,
		authority:  authority,

		validatorAddressCodec: validatorAddressCodec,

		NextBridgeId:      collections.NewSequence(sb, types.NextBridgeIdKey, "next_bridge_id"),
		Params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		BridgeConfigs:     collections.NewMap(sb, types.BridgeConfigPrefix, "bridge_configs", collections.Uint64Key, codec.CollValue[types.BridgeConfig](cdc)),
		BatchInfos:        collections.NewMap(sb, types.BatchInfoPrefix, "batch_infos", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), codec.CollValue[types.BatchInfoWithOutput](cdc)),
		NextL1Sequences:   collections.NewMap(sb, types.NextL1SequencePrefix, "next_l1_sequences", collections.Uint64Key, collections.Uint64Value),
		TokenPairs:        collections.NewMap(sb, types.TokenPairPrefix, "token_pairs", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), collections.StringValue),
		OutputProposals:   collections.NewMap(sb, types.OutputProposalPrefix, "output_proposals", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), codec.CollValue[types.Output](cdc)),
		NextOutputIndexes: collections.NewMap(sb, types.NextOutputIndexPrefix, "next_output_indexes", collections.Uint64Key, collections.Uint64Value),
		ProvenWithdrawals: collections.NewMap(sb, types.ProvenWithdrawalPrefix, "proven_withdrawals", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), collections.BoolValue),
		MigrationInfos:    collections.NewMap(sb, types.MigrationInfoPrefix, "migration_infos", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), codec.CollValue[types.MigrationInfo](cdc)),
		OraclePriceHash:   collections.NewItem(sb, types.OraclePriceHashPrefix, "oracle_price_hash", codec.CollValue[types.OraclePriceHash](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the x/move module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ValidatorAddressCodec returns the validator address codec.
func (k Keeper) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Codec returns the keeper codec
func (k Keeper) Codec() codec.Codec {
	return k.cdc
}

// IsBound and BindPort are no-op keeping for backward compatibility with the
// v1.4.5 upgrade handler in initia, which called these before ibc-go v10 dropped
// the capability module. Port binding now happens via router registration.
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	return true
}

func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	return nil
}
