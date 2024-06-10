package steps

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	conmetconfig "github.com/cometbft/cometbft/config"
	cometos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/go-bip39"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	"github.com/pkg/errors"
)

type (
	GenesisPostSetupFunc func(
		launchtools.Launcher,
		*launchtools.Config,
		map[string]json.RawMessage,
		codec.Codec,
	) (map[string]json.RawMessage, error)
)

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = InitializeGenesis

// InitializeGenesis initializes the genesis state for the application.
// Note: if you prefer adding more to the genesis, consider using InitializeGenesisWithPostSetup.
func InitializeGenesis(
	config *launchtools.Config,
) launchtools.LauncherStepFunc {
	return InitializeGenesisWithPostSetup()(config)
}

// InitializeGenesisWithPostSetup initializes the genesis state for the application.
// This function accepts a list of post-setup functions that can be used to modify the genesis state.
func InitializeGenesisWithPostSetup(
	postsetup ...GenesisPostSetupFunc,
) launchtools.LauncherStepFuncFactory[*launchtools.Config] {
	return func(config *launchtools.Config) launchtools.LauncherStepFunc {
		return func(ctx launchtools.Launcher) error {
			if ctx.IsAppInitialized() {
				return errors.New("application already initialized. InitializeGenesis should only be called once")
			}

			appGenesis, err := initializeGenesis(
				ctx,
				config,
				ctx.Logger(),
				ctx.ClientContext().Codec,
				ctx.ServerContext().Config,
				ctx.DefaultGenesis(),
				postsetup...,
			)

			if err != nil {
				return errors.Wrap(err, "failed to initialize genesis")
			}

			// store genesis
			if err := genutil.ExportGenesisFile(appGenesis, ctx.ServerContext().Config.GenesisFile()); err != nil {
				return errors.Wrap(err, "failed to export genesis file")
			}

			return nil
		}
	}
}

func initializeGenesis(
	ctx launchtools.Launcher,
	config *launchtools.Config,
	log log.Logger,
	cdc codec.Codec,
	cometConfig *conmetconfig.Config,
	genesisAppState map[string]json.RawMessage,
	postsetup ...GenesisPostSetupFunc,
) (*genutiltypes.AppGenesis, error) {
	log.Info("initializing genesis")

	// create validator mnemonic for sequencer operation
	validatorKeySpec := config.SystemKeys.Validator
	if !bip39.IsMnemonicValid(validatorKeySpec.Mnemonic) {
		return nil, errors.New("invalid mnemonic for validator key")
	}

	// initialize default configs with validator system key.
	// this must succeed, given validatorKeySpec is pre-validated
	nodeId, valPubKey, err := genutil.InitializeNodeValidatorFilesFromMnemonic(
		cometConfig,
		validatorKeySpec.Mnemonic,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize node validator files")
	}

	log.Info(
		"created node identity",
		"node_id", nodeId,
		"chain_id", config.L2Config.ChainID,
		"validator_address", validatorKeySpec.Address,
		"moniker", cometConfig.Moniker,
	)

	// prepare genesis
	genFilePath := cometConfig.GenesisFile()
	if cometos.FileExists(genFilePath) {
		return nil, errors.Wrap(err, "genesis file already exists")
	}

	// prepare default genesis
	// reuse whatever the default genesis generator is
	// then add parts that require to be part of initial genesis
	// such as sequence validator

	// Step 1 -------------------------------------------------------------------------------------------
	// Add genesis accounts to auth and bank modules
	// iterate over all GenesisAccounts from config, validate them, and add them to the genesis state.
	// this call modifies appstate.auth, appstate.bank
	log.Info("adding genesis accounts", "accounts-len", len(*config.GenesisAccounts))
	genesisAuthState, genesisBankState, err := addGenesisAccounts(
		cdc,
		genesisAppState,
		*config.GenesisAccounts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add genesis accounts")
	}

	genesisAppState[authtypes.ModuleName] = cdc.MustMarshalJSON(genesisAuthState)
	genesisAppState[banktypes.ModuleName] = cdc.MustMarshalJSON(genesisBankState)

	// Step 2 -------------------------------------------------------------------------------------------
	// Add genesis validator to opchild module
	// this call modifies appstate.opchild
	log.Info("adding genesis validator",
		"moniker", config.L2Config.Moniker,
		"validator_address_acc", validatorKeySpec.Address,
		"validator_address_val", sdk.ValAddress(valPubKey.Address()).String(),
	)
	opChildState, err := addGenesisValidator(
		cdc,
		genesisAppState,
		config.L2Config.Moniker,
		valPubKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add genesis validator")
	}

	genesisAppState[opchildtypes.ModuleName] = cdc.MustMarshalJSON(opChildState)

	// Step 3 -------------------------------------------------------------------------------------------
	// Add fee whitelist to genesis
	// whitelist specific operators for fee exemption
	log.Info("adding fee whitelists",
		"whitelist-len", 3,
		"whitelists", strings.Join([]string{
			config.SystemKeys.Validator.Address,
			config.SystemKeys.BridgeExecutor.Address,
			config.SystemKeys.Challenger.Address,
		}, ","),
	)
	opChildState, err = addFeeWhitelists(cdc, genesisAppState, []string{
		config.SystemKeys.Validator.Address,
		config.SystemKeys.BridgeExecutor.Address,
		config.SystemKeys.Challenger.Address,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to add fee whitelists")
	}

	genesisAppState[opchildtypes.ModuleName] = cdc.MustMarshalJSON(opChildState)

	// Step 4 -------------------------------------------------------------------------------------------
	// Set bridge executor address in the genesis parameter
	log.Info("setting bridge executor address",
		"bridge-executor", config.SystemKeys.BridgeExecutor.Address,
	)

	opChildState, err = setOpChildBridgeExecutorAddress(cdc, genesisAppState, config.SystemKeys.BridgeExecutor.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set bridge executor address")
	}

	genesisAppState[opchildtypes.ModuleName] = cdc.MustMarshalJSON(opChildState)

	// Step 5 -------------------------------------------------------------------------------------------
	// Set admin address in the genesis parameter
	log.Info("setting admin address",
		"admin", config.SystemKeys.Validator.Address,
	)

	opChildState, err = setOpChildAdminAddress(cdc, genesisAppState, config.SystemKeys.Validator.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set bridge executor address")
	}

	genesisAppState[opchildtypes.ModuleName] = cdc.MustMarshalJSON(opChildState)

	// run post-setup, if any
	for _, setup := range postsetup {
		genesisAppState, err = setup(ctx, config, genesisAppState, cdc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to run post-setup")
		}
	}

	// finalize app genesis
	appGenesis := &genutiltypes.AppGenesis{}
	appGenesis.Consensus = &genutiltypes.ConsensusGenesis{
		Validators: nil,
	}
	appGenesis.AppName = version.AppName
	appGenesis.AppVersion = version.Version
	appGenesis.ChainID = config.L2Config.ChainID
	appGenesis.AppState, err = json.MarshalIndent(genesisAppState, "", " ")
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal app state")
	}

	// validate genesis
	if err := appGenesis.ValidateAndComplete(); err != nil {
		return nil, errors.Wrap(err, "failed to validate and complete app genesis")
	}

	return appGenesis, nil
}

func addGenesisAccounts(cdc codec.Codec, genesisAppState map[string]json.RawMessage, genAccsManifest []launchtools.AccountWithBalance) (
	*authtypes.GenesisState,
	*banktypes.GenesisState,
	error,
) {
	// handle adding genesis accounts to auth and bank state
	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	authGenState := authtypes.GetGenesisStateFromAppState(cdc, genesisAppState)

	// iterate over all genesis accounts from config, validate them, and add them to the genesis state
	authAccounts, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get accounts from genesis state")
	}

	for _, acc := range genAccsManifest {
		// acc
		addr, addrErr := sdk.AccAddressFromBech32(acc.Address)
		if addrErr != nil {
			return nil, nil, errors.Wrap(addrErr, fmt.Sprintf("failed to parse genesis account address %s", acc.Address))
		}

		genAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
		if err := genAccount.Validate(); err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("failed to validate genesis account: %s", acc.Address))
		}
		authAccounts = append(authAccounts, genAccount)

		// bank
		coins, err := sdk.ParseCoinsNormalized(acc.Coins)
		if err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("failed to parse genesis account coins: %s", acc.Address))
		}
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
			Address: acc.Address,
			Coins:   coins.Sort(),
		})
	}

	// convert accounts into any's
	genesisAccounts, err := authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(authAccounts))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to convert accounts into any's")
	}

	authGenState.Accounts = genesisAccounts
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	return &authGenState, bankGenState, nil
}

func addGenesisValidator(
	cdc codec.Codec,
	genesisAppState map[string]json.RawMessage,
	moniker string,
	valPubKey cryptotypes.PubKey,
) (
	*opchildtypes.GenesisState,
	error,
) {
	valAddr := sdk.ValAddress(valPubKey.Address())
	validator, err := opchildtypes.NewValidator(valAddr, valPubKey, moniker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create genesis validator")
	}

	// inscribe the validator into the genesis state
	opChildState := opchildtypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	opChildState.Validators = append((*opChildState).Validators, validator)

	return opChildState, nil
}

func addFeeWhitelists(cdc codec.Codec, genesisAppState map[string]json.RawMessage, whitelistAddrs []string) (
	*opchildtypes.GenesisState,
	error,
) {
	opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	opchildState.Params.FeeWhitelist = append(opchildState.Params.FeeWhitelist, whitelistAddrs...)

	return opchildState, nil
}

func setOpChildAdminAddress(cdc codec.Codec, genesisAppState map[string]json.RawMessage, adminAddr string) (
	*opchildtypes.GenesisState,
	error,
) {
	opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	opchildState.Params.Admin = adminAddr

	return opchildState, nil
}

func setOpChildBridgeExecutorAddress(cdc codec.Codec, genesisAppState map[string]json.RawMessage, bridgeExecutorAddr string) (
	*opchildtypes.GenesisState,
	error,
) {
	opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	opchildState.Params.BridgeExecutors = []string{bridgeExecutorAddr}

	return opchildState, nil
}

func setMultiOpChildBridgeExecutorsAddress(cdc codec.Codec, genesisAppState map[string]json.RawMessage, bridgeExecutorsAddrs []string) (
	*opchildtypes.GenesisState,
	error,
) {
	opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, genesisAppState)
	opchildState.Params.BridgeExecutors = bridgeExecutorsAddrs

	return opchildState, nil
}
