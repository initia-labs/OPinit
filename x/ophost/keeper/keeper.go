package keeper

import (
	"context"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeService corestoretypes.KVStoreService

	authKeeper          types.AccountKeeper
	bankKeeper          types.BankKeeper
	bridgeHook          types.BridgeHook
	communityPoolKeeper types.CommunityPoolKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

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
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.CommunityPoolKeeper,
	bridgeHook types.BridgeHook,
	authority string,
) *Keeper {
	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := &Keeper{
		cdc:          cdc,
		storeService: storeService,

		authKeeper:          ak,
		bankKeeper:          bk,
		communityPoolKeeper: ck,

		bridgeHook: bridgeHook,
		authority:  authority,

		NextBridgeId:      collections.NewSequence(sb, types.NextBridgeIdKey, "next_bridge_id"),
		Params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		BridgeConfigs:     collections.NewMap(sb, types.BridgeConfigPrefix, "bridge_configs", collections.Uint64Key, codec.CollValue[types.BridgeConfig](cdc)),
		BatchInfos:        collections.NewMap(sb, types.BatchInfoPrefix, "batch_infos", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), codec.CollValue[types.BatchInfoWithOutput](cdc)),
		NextL1Sequences:   collections.NewMap(sb, types.NextL1SequencePrefix, "next_l1_sequences", collections.Uint64Key, collections.Uint64Value),
		TokenPairs:        collections.NewMap(sb, types.TokenPairPrefix, "token_pairs", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), collections.StringValue),
		OutputProposals:   collections.NewMap(sb, types.OutputProposalPrefix, "output_proposals", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), codec.CollValue[types.Output](cdc)),
		NextOutputIndexes: collections.NewMap(sb, types.NextOutputIndexPrefix, "next_output_indexes", collections.Uint64Key, collections.Uint64Value),
		ProvenWithdrawals: collections.NewMap(sb, types.ProvenWithdrawalPrefix, "proven_withdrawals", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), collections.BoolValue),
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

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}
