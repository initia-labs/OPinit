package types

import (
	context "context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

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
}

type OracleKeeper interface {
	currencypair.OracleKeeper
	GetAllCurrencyPairs(ctx sdk.Context) []connecttypes.CurrencyPair
	SetPriceForCurrencyPair(ctx sdk.Context, cp connecttypes.CurrencyPair, qp oracletypes.QuotePrice) error
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
