package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/initia-labs/OPinit/x/opchild"
	"github.com/initia-labs/OPinit/x/opchild/client/cli"
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
	s.encCfg = testutilmod.MakeTestEncodingConfig(opchild.AppModuleBasic{}, bank.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen().WithOutput(&outBuf)

	s.addrs = make([]sdk.AccAddress, 0)
	for i := 0; i < 3; i++ {
		k, _, err := s.clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)

		pub, err := k.GetPubKey()
		s.Require().NoError(err)

		newAddr := sdk.AccAddress(pub.Address())
		s.addrs = append(s.addrs, newAddr)
	}
}

func (s *CLITestSuite) TestNewWithdrawCmd() {
	require := s.Require()
	cmd := cli.NewWithdrawCmd(addresscodec.NewBech32Codec("init"))

	consPrivKey := ed25519.GenPrivKey()
	consPubKeyBz, err := s.encCfg.Codec.MarshalInterfaceJSON(consPrivKey.PubKey())
	require.NoError(err)
	require.NotNil(consPubKeyBz)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid to_l1)",
			[]string{
				"_invalid_acc_",
				"100umin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid amount)",
			[]string{
				s.addrs[0].String(),
				"0umin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
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

func (s *CLITestSuite) TestNewDepositCmd() {
	require := s.Require()
	cmd := cli.NewDepositCmd(addresscodec.NewBech32Codec("init"))

	consPrivKey := ed25519.GenPrivKey()
	consPubKeyBz, err := s.encCfg.Codec.MarshalInterfaceJSON(consPrivKey.PubKey())
	require.NoError(err)
	require.NotNil(consPubKeyBz)

	hookMsg := "hello world"
	hookMsgBytes, err := json.Marshal(hookMsg)
	require.NoError(err)
	hookMsgString := string(hookMsgBytes)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (invalid from_l1)",
			[]string{
				"1",
				"_invalid_acc_",
				s.addrs[0].String(),
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid to_l2)",
			[]string{
				s.addrs[0].String(),
				"_invalid_acc_",
				"1",
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid sequence)",
			[]string{
				"-1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid amount)",
			[]string{
				"1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"0umin",
				"test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid base_denom)",
			[]string{
				"1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"100umin",
				"1test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction without hook msg",
			[]string{
				"1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction with valid hook msg",
			[]string{
				"1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", cli.FlagHookMsg, hookMsgString),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction with invalid hook msg",
			[]string{
				"1",
				s.addrs[0].String(),
				s.addrs[1].String(),
				"100umin",
				"test_token",
				fmt.Sprintf("--%s=%s", cli.FlagHookMsg, "__invalid_hook_msg__"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
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

func (s *CLITestSuite) TestNewExecuteMessagesCmd() {
	require := s.Require()
	cmd := cli.NewExecuteMessagesCmd(addresscodec.NewBech32Codec("init"))

	consPrivKey := ed25519.GenPrivKey()
	consPubKeyBz, err := s.encCfg.Codec.MarshalInterfaceJSON(consPrivKey.PubKey())
	require.NoError(err)
	require.NotNil(consPubKeyBz)

	emptyMessages := `{"messages": []}`
	emptyMessagesFile := testutil.WriteToNewTempFile(s.T(), emptyMessages)
	defer emptyMessagesFile.Close()

	invalidMessages := `{"messages": [
    {
      "@type": "/cosmos.bank.v1beta1.InvalidMsgSend",
      "from_address": "inval1...",
      "to_address": "inval1...",
      "amount":[{"denom": "uinval","amount": "inval"}]
    }
    ]}`
	invalidMessagesFile := testutil.WriteToNewTempFile(s.T(), invalidMessages)
	defer invalidMessagesFile.Close()

	validMessages := `{"messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "init12jea62yu8tmhzy7zage0fngyjvsskfyg4n9y34",
      "to_address": "init15hqqzhye49577x45ekz24jhesvc442zzamxksx",
      "amount":[{"denom": "umin","amount": "10"}]
    }
    ]}`
	validMessagesFile := testutil.WriteToNewTempFile(s.T(), validMessages)
	defer validMessagesFile.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid transaction (empty messages)",
			[]string{
				emptyMessagesFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"invalid transaction (invalid in-messages)",
			[]string{
				invalidMessagesFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10))).String()),
			},
			true, 0, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			[]string{
				validMessagesFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addrs[0]),
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
