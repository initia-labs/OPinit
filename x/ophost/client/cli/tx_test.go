package cli_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	math "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/initia-labs/OPinit/x/ophost"
	"github.com/initia-labs/OPinit/x/ophost/client/cli"
)

var PKs = simtestutil.CreateTestPubKeys(500)

type CLITestSuite struct {
	suite.Suite

	ac        address.Codec
	kr        keyring.Keyring
	encCfg    testutilmod.TestEncodingConfig
	baseCtx   client.Context
	clientCtx client.Context
	addrs     []sdk.AccAddress
}

func (s *CLITestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("init", "initpub")
	s.ac = addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	s.encCfg = testutilmod.MakeTestEncodingConfig(ophost.AppModuleBasic{}, bank.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	s.addrs = make([]sdk.AccAddress, 0)
	for i := 0; i < 3; i++ {
		k, _, err := s.clientCtx.Keyring.NewMnemonic(fmt.Sprintf("NewValidator%d", i), keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.NoError(err)

		addr, err := k.GetAddress()
		s.NoError(err)
		s.addrs = append(s.addrs, addr)
	}
}

func (s *CLITestSuite) TestNewRecordBatchCmd() {
	require := s.Require()
	cmd := cli.NewRecordBatchCmd(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	s.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid bridge_id)",
			[]string{
				"0",
				"Ynl0ZXM=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid batch_bytes)",
			[]string{
				"1",
				"batch_bytes_should_be_base64_encoded",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				"1",
				"Ynl0ZXM=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

//nolint:dupl
func (s *CLITestSuite) TestNewCreateBridge() {
	require := s.Require()
	cmd := cli.NewCreateBridge(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	s.NoError(err)

	invalidConfig, err := os.CreateTemp("/tmp", "bridge_config")
	require.NoError(err)
	defer os.Remove(invalidConfig.Name())
	validConfig, err := os.CreateTemp("/tmp", "bridge_config")
	require.NoError(err)
	defer os.Remove(validConfig.Name())

	_, err = invalidConfig.WriteString(`{}`)
	s.NoError(err)
	_, err = validConfig.WriteString(`{
        "challenger": "init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
        "proposer": "init1k2svyvm60r8rhnzr9vemk5f6fksvm6tyeujp3c",
        "submission_interval": "100s",
        "finalization_period": "1000s",
        "submission_start_height" : "1",
        "metadata": "channel-0",
		"batch_info": {
			"submitter": "init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
			"chain_type": "CELESTIA"
		},
		"oracle_enabled": true
    }`)
	s.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid bridge config)",
			[]string{
				invalidConfig.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				validConfig.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewProposeOutput() {
	require := s.Require()
	cmd := cli.NewProposeOutput(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	s.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid bridge-id)",
			[]string{
				"0",
				"1",
				"1234",
				"XK+A/MAyMz3pzUNsVNqKNPPQh7Up64VsRTycHqA8arE=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid l2-block-nmber)",
			[]string{
				"1",
				"1",
				"-1",
				"XK+A/MAyMz3pzUNsVNqKNPPQh7Up64VsRTycHqA8arE=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid output-root-hash)",
			[]string{
				"1",
				"1",
				"1234",
				"--XK+A/MAyMz3pzUNsVNqKNPPQh7Up64VsRTycHqA8arE=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				"1",
				"1",
				"1234",
				"XK+A/MAyMz3pzUNsVNqKNPPQh7Up64VsRTycHqA8arE=",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewDeleteOutput() {
	require := s.Require()

	cmd := cli.NewDeleteOutput(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	s.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid bridge-id)",
			[]string{
				"0",
				"1000",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid output-index)",
			[]string{
				"1",
				"-1",
				"2e297e695e451144fc44db083d6b3d56f0a5f920721e3efc90ec7662c7775d1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				"1",
				"2",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func (s *CLITestSuite) TestNewInitiateTokenDeposit() {
	require := s.Require()
	cmd := cli.NewInitiateTokenDeposit(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	require.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid bridge-id)",
			[]string{
				"0",
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				"10000uatom",
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid to)",
			[]string{
				"0",
				"cosmos1q6jhwnarkw2j5qqgx3qlu20k8nrdglft6qssy3",
				"10000uatom",
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid amount)",
			[]string{
				"0",
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				"invalid_amount",
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid data)",
			[]string{
				"0",
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				"10000uatom",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				"1",
				"init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
				"10000uatom",
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

//nolint:dupl
func (s *CLITestSuite) TestNewFinalizeTokenWithdrawal() {
	require := s.Require()
	cmd := cli.NewFinalizeTokenWithdrawal(s.ac)

	addr0, err := s.ac.BytesToString(s.addrs[0])
	s.NoError(err)

	invalidConfig, err := os.CreateTemp("/tmp", "withdrawal_info")
	require.NoError(err)
	defer os.Remove(invalidConfig.Name())
	validConfig, err := os.CreateTemp("/tmp", "withdrawal_info")
	require.NoError(err)
	defer os.Remove(validConfig.Name())

	_, err = invalidConfig.WriteString(`{}`)
	s.NoError(err)
	_, err = validConfig.WriteString(`{
		"bridge_id": "1",
		"output_index": "1180",
		"withdrawal_proofs": [
			"q6T8JJm7AdbD4rgZ3BjanRHdE1x7aLZwp36pPrOOey4="
		],
		"from": "init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
		"to": "init1q6jhwnarkw2j5qqgx3qlu20k8nrdglft5ksr0g",
		"sequence": "5",
		"amount": {"amount": "100", "denom": "uinit"},
		"version": "AQ==",
		"storage_root": "KGlalV+mBHC7YFOLNX3g9LLzmyvP7QCm42HKo9N3Lu8=",
		"last_block_hash": "6oFdc+PEkXVJAo5IpXJ91vbCT9FNuKCz5VSlaFmxG+Y="
		}`)
	s.NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid withdrawal info)",
			[]string{
				invalidConfig.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				validConfig.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, addr0),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
			} else {
				require.NoError(err, "test: %s\noutput: %s", tc.name, out.String())
				err = s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(err, out.String(), "test: %s, output\n:", tc.name, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				require.Equal(tc.expectedCode, txResp.Code,
					"test: %s, output\n:", tc.name, out.String())
			}
		})
	}
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}
