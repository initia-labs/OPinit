package ante_test

import (
	"context"
	"testing"

	"github.com/initia-labs/OPinit/v1/x/opchild/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	errorsmod "cosmossdk.io/errors"

	"github.com/initia-labs/OPinit/v1/x/opchild/ante"
)

// AnteTestSuite is a test suite to be used with ante handler tests.
type AnteTestSuite struct {
	suite.Suite

	app       *simapp.SimApp
	ctx       sdk.Context
	clientCtx client.Context
	txBuilder client.TxBuilder
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool, tempDir string) (*simapp.SimApp, sdk.Context) {
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = tempDir
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := simapp.NewSimApp(
		log.NewNopLogger(), dbm.NewMemDB(), nil, true, appOptions,
	)
	ctx := app.BaseApp.NewContext(isCheckTx)
	err := app.AccountKeeper.Params.Set(ctx, authtypes.DefaultParams())
	if err != nil {
		panic(err)
	}

	return app, ctx
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (suite *AnteTestSuite) SetupTest(isCheckTx bool) {
	tempDir := suite.T().TempDir()
	suite.app, suite.ctx = createTestApp(isCheckTx, tempDir)
	suite.ctx = suite.ctx.WithBlockHeight(1)

	// We're using TestMsg encoding in some tests, so register it here.
	suite.app.LegacyAmino().RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(suite.app.InterfaceRegistry())

	suite.clientCtx = client.Context{}.
		WithTxConfig(suite.app.TxConfig())
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(
	ctx context.Context, privs []cryptotypes.PrivKey,
	accNums, accSeqs []uint64,
	chainID string, signMode signing.SignMode,
) (xauthsigning.Tx, error) {
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err := suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			ctx, signMode, signerData,
			suite.txBuilder, priv, suite.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = suite.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return suite.txBuilder.GetTx(), nil
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

var _ sdk.Tx = testTx{}

type testTx struct {
	msgs []sdk.Msg
}

func (t testTx) GetMsgsV2() ([]protoreflect.ProtoMessage, error) {
	return nil, nil
}

func (t testTx) GetMsgs() []sdk.Msg {
	return t.msgs
}

func TestRedundantTx(t *testing.T) {
	ctx, input := createTestInput(t, true)
	rbd := ante.NewRedundantBridgeDecorator(&input.OPChildKeeper)

	// input.Faucet.Mint(ctx, addrs[0], sdk.NewCoin(testDenoms[0], math.NewInt(100000)))

	tx := testTx{
		msgs: []sdk.Msg{
			types.NewMsgFinalizeTokenDeposit(
				addrsStr[0], addrsStr[0], addrsStr[1], sdk.NewCoin(testDenoms[0], math.NewInt(100)), 1, 1, "l1_test0", nil,
			),
			types.NewMsgFinalizeTokenDeposit(
				addrsStr[0], addrsStr[0], addrsStr[1], sdk.NewCoin(testDenoms[0], math.NewInt(100)), 2, 1, "l1_test0", nil,
			),
		},
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_, err := rbd.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return sdk.Context{}, nil })
	require.NoError(t, err)

	_, err = rbd.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return sdk.Context{}, nil })
	require.True(t, errorsmod.IsOf(err, types.ErrRedundantTx))

	tx = testTx{
		msgs: []sdk.Msg{
			types.NewMsgFinalizeTokenDeposit(
				addrsStr[0], addrsStr[0], addrsStr[1], sdk.NewCoin(testDenoms[0], math.NewInt(100)), 2, 1, "l1_test0", nil,
			),
			types.NewMsgFinalizeTokenDeposit(
				addrsStr[0], addrsStr[0], addrsStr[1], sdk.NewCoin(testDenoms[0], math.NewInt(100)), 3, 1, "l1_test0", nil,
			),
		},
	}

	_, err = rbd.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return sdk.Context{}, nil })
	require.NoError(t, err)

	tx = testTx{
		msgs: []sdk.Msg{
			types.NewMsgFinalizeTokenDeposit(
				addrsStr[0], addrsStr[0], addrsStr[1], sdk.NewCoin(testDenoms[0], math.NewInt(100)), 4, 1, "l1_test0", nil,
			),
		},
	}

	_, err = rbd.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) { return sdk.Context{}, nil })
	require.NoError(t, err)
}
