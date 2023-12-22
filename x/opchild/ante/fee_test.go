package ante_test

import (
	"github.com/initia-labs/OPinit/x/opchild/ante"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var baseDenom = sdk.DefaultBondDenom

type TestAnteKeeper struct {
	minGasPrices sdk.DecCoins
}

func (k TestAnteKeeper) MinGasPrices(ctx sdk.Context) sdk.DecCoins {
	return k.minGasPrices
}

func (suite *AnteTestSuite) TestEnsureMempoolFees() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	// set price 0.5 base == 1 quote
	fc := ante.NewMempoolFeeChecker(TestAnteKeeper{
		minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyZeroDec())),
	})

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	// gas price 0.0005
	msg := testdata.NewTestMsg(addr1)
	feeAmount := sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(100)))
	gasLimit := uint64(200_000)

	suite.Require().NoError(suite.txBuilder.SetMsgs(msg))
	suite.txBuilder.SetFeeAmount(feeAmount)
	suite.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := suite.CreateTestTx(suite.ctx, privs, accNums, accSeqs, suite.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	suite.Require().NoError(err)

	// Set high gas price so standard test fee fails
	basePrice := sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(200).Quo(math.LegacyNewDec(100000)))
	highGasPrice := []sdk.DecCoin{basePrice}
	suite.ctx = suite.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, _, err = fc.CheckTxFeeWithMinGasPrices(suite.ctx, tx)
	suite.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	suite.ctx = suite.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, _, err = fc.CheckTxFeeWithMinGasPrices(suite.ctx, tx)
	suite.Require().Nil(err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	// gas price = 0.0005
	basePrice = sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDecWithPrec(5, 4))
	lowGasPrice := []sdk.DecCoin{basePrice}
	suite.ctx = suite.ctx.WithMinGasPrices(lowGasPrice)

	_, _, err = fc.CheckTxFeeWithMinGasPrices(suite.ctx, tx)
	suite.Require().Nil(err, "Decorator should not have errored on fee higher than local gasPrice")

	// set high base_min_gas_price to test should be failed
	fc = ante.NewMempoolFeeChecker(TestAnteKeeper{
		minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(200).Quo(math.LegacyNewDec(100000)))),
	})

	suite.txBuilder.SetFeeAmount(feeAmount)
	_, _, err = fc.CheckTxFeeWithMinGasPrices(suite.ctx, tx)
	suite.Require().NotNil(err, "Decorator should have errored on too low fee for local gasPrice")
}

func (suite *AnteTestSuite) TestCombinedMinGasPrices() {
	minGasPrices := sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(100)))
	configMinGasPrices := sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(100)))

	combined := ante.CombinedMinGasPrices(minGasPrices, configMinGasPrices)
	suite.Require().Len(combined, 1)
	suite.Require().True(combined.AmountOf(baseDenom).Equal(math.LegacyNewDec(100)))

	configMinGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(150)))
	combined = ante.CombinedMinGasPrices(minGasPrices, configMinGasPrices)
	suite.Require().Len(combined, 1)
	suite.Require().True(combined.AmountOf(baseDenom).Equal(math.LegacyNewDec(150)))

	minGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(200)))
	combined = ante.CombinedMinGasPrices(minGasPrices, configMinGasPrices)
	suite.Require().Len(combined, 1)
	suite.Require().True(combined.AmountOf(baseDenom).Equal(math.LegacyNewDec(200)))

	configMinGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(100)), sdk.NewDecCoinFromDec("test", math.LegacyNewDec(100)))
	combined = ante.CombinedMinGasPrices(minGasPrices, configMinGasPrices)
	suite.Require().Len(combined, 2)
	suite.Require().True(combined.AmountOf(baseDenom).Equal(math.LegacyNewDec(200)))
	suite.Require().True(combined.AmountOf("test").Equal(math.LegacyNewDec(100)))

	minGasPrices = sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom, math.LegacyNewDec(100)), sdk.NewDecCoinFromDec("test2", math.LegacyNewDec(300)))
	combined = ante.CombinedMinGasPrices(minGasPrices, configMinGasPrices)
	suite.Require().Len(combined, 3)
	suite.Require().True(combined.AmountOf(baseDenom).Equal(math.LegacyNewDec(100)))
	suite.Require().True(combined.AmountOf("test").Equal(math.LegacyNewDec(100)))
	suite.Require().True(combined.AmountOf("test2").Equal(math.LegacyNewDec(300)))
}
