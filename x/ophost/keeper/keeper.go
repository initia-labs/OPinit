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
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

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
	portKeeper          types.PortKeeper
	scopedKeeper        types.ScopedKeeper

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
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	msgRouter baseapp.MessageRouter,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.CommunityPoolKeeper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper types.ScopedKeeper,
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
		portKeeper:          portKeeper,
		scopedKeeper:        scopedKeeper,

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

// IsBound checks if the ophost module is already bound to the desired port
func (k Keeper) IsBound(ctx context.Context, portID string) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, ok := k.scopedKeeper.GetCapability(sdkCtx, host.PortPath(portID))

	return ok
}

// BindPort defines a wrapper function for the Keeper's function to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx context.Context, portID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	capability := k.portKeeper.BindPort(sdkCtx, portID)

	return k.ClaimCapability(ctx, capability, host.PortPath(portID))
}

// ClaimCapability claims a channel capability for the ophost module
func (k Keeper) ClaimCapability(ctx context.Context, cap *capabilitytypes.Capability, name string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return k.scopedKeeper.ClaimCapability(sdkCtx, cap, name)
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
