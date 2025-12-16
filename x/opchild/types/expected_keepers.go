package types

import (
	context "context"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	AddressCodec() address.Codec
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI

	IterateAccounts(ctx context.Context, process func(sdk.AccountI) (stop bool))
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI // only used for simulation
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	SetAccount(ctx context.Context, acc sdk.AccountI)

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	GetSupply(ctx context.Context, denom string) sdk.Coin

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderPool, recipientPool string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error

	HasDenomMetaData(ctx context.Context, denom string) bool
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)

	IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool)
}

type OracleKeeper interface {
	currencypair.OracleKeeper
	SetPriceForCurrencyPair(ctx context.Context, cp connecttypes.CurrencyPair, qp oracletypes.QuotePrice) error
}

type TransferKeeper interface {
	DenomPathFromHash(ctx sdk.Context, denom string) (string, error)
}

type ChannelKeeper interface {
	GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error)
}

// ValidatorSet expected properties for the set of all validators (noalias)
type ValidatorSet interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(context.Context, func(validator ValidatorI) (stop bool, err error)) error

	Validator(context.Context, sdk.ValAddress) ValidatorI            // get a particular validator by operator address
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) ValidatorI // get a particular validator by consensus address

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(context.Context) (uint32, error)
}

// ClientKeeper defines the expected IBC client keeper interface for proof verification
type ClientKeeper interface {
	// GetClientState returns the client state for the given client ID
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	// ClientStore returns the client-prefixed store for the given client ID
	ClientStore(ctx sdk.Context, clientID string) storetypes.KVStore
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// ScopedKeeper defines the expected IBC scoped keeper for managing channel capabilities
type ScopedKeeper interface {
	GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool)
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
}
