package cli

import (
	"bufio"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/address"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	opchildtypes "github.com/initia-labs/OPinit/v1/x/opchild/types"
)

// AddGenesisValidatorCmd builds the application's add-genesis-validator command.
func AddGenesisValidatorCmd(mbm module.BasicManager, txEncCfg client.TxEncodingConfig, genBalIterator genutiltypes.GenesisBalancesIterator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-validator [key_name]",
		Short: "Add a genesis validator",
		Args:  cobra.ExactArgs(1),
		Long: fmt.Sprintf(`Add a genesis validator with the key in the Keyring referenced by a given name.
		A Bech32 consensus pubkey may optionally be provided.

Example:
$ %s add-genesis-validator my-key-name --home=/path/to/home/dir --keyring-backend=os --chain-id=test-chain-1
`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			_ /*nodeId*/, valPubKey, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}

			// read --pubkey, if empty take it from priv_validator.json
			if pkStr, _ := cmd.Flags().GetString(FlagPubKey); pkStr != "" {
				if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &valPubKey); err != nil {
					return errors.Wrap(err, "failed to unmarshal validator public key")
				}
			}

			name := args[0]
			key, err := clientCtx.Keyring.Key(name)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch '%s' from the keyring", name)
			}

			moniker := config.Moniker
			if m, _ := cmd.Flags().GetString(FlagMoniker); m != "" {
				moniker = m
			}

			addr, err := key.GetAddress()
			if err != nil {
				return err
			}
			valAddr := sdk.ValAddress(addr)

			validator, err := opchildtypes.NewValidator(valAddr, valPubKey, moniker)
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, appState)
			opchildState.Validators = append((*opchildState).Validators, validator)
			if opchildState.Params.Admin == "" {
				opchildState.Params.Admin = addr.String()
			}

			allEmpty := true
			for _, be := range opchildState.Params.BridgeExecutors {
				if be != "" {
					allEmpty = false
					break
				}
			}
			if allEmpty {
				opchildState.Params.BridgeExecutors = []string{addr.String()}
			}

			opchildGenStateBz, err := cdc.MarshalJSON(opchildState)
			if err != nil {
				return fmt.Errorf("failed to marshal opchild genesis state: %w", err)
			}
			appState[opchildtypes.ModuleName] = opchildGenStateBz

			if err = mbm.ValidateGenesis(cdc, txEncCfg, appState); err != nil {
				return errors.Wrap(err, "failed to validate genesis state")
			}
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			if err = genutil.ExportGenesisFile(genDoc, config.GenesisFile()); err != nil {
				return errors.New("Failed to export genesis file")
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AddFeeWhitelistCmd builds the application's fee-whitelist command.
func AddFeeWhitelistCmd(defaultNodeHome string, addressCodec address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-fee-whitelist [address_or_key_name]",
		Short: "Add an address to the fee whitelist",
		Args:  cobra.ExactArgs(1),
		Long: `Add an address to fee whitelist of genesis.json. The provided account must specify
the account address or key name . If a key name is given,
the address will be looked up in the local Keybase.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			var kr keyring.Keyring
			addr, err := addressCodec.StringToBytes(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

				if keyringBackend != "" && clientCtx.Keyring == nil {
					var err error
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, clientCtx.Codec)
					if err != nil {
						return err
					}
				} else {
					kr = clientCtx.Keyring
				}

				k, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring: %w", err)
				}

				addr, err = k.GetAddress()
				if err != nil {
					return err
				}
			}
			addrStr, err := addressCodec.BytesToString(addr)
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			cdc := clientCtx.Codec
			opchildState := opchildtypes.GetGenesisStateFromAppState(cdc, appState)
			opchildState.Params.FeeWhitelist = append(opchildState.Params.FeeWhitelist, addrStr)

			opchildGenStateBz, err := cdc.MarshalJSON(opchildState)
			if err != nil {
				return fmt.Errorf("failed to marshal opchild genesis state: %w", err)
			}
			appState[opchildtypes.ModuleName] = opchildGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			if err = genutil.ExportGenesisFile(genDoc, config.GenesisFile()); err != nil {
				return errors.New("Failed to export genesis file")
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
