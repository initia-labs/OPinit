package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/initia-labs/OPinit/x/ophost/keeper"
	"github.com/initia-labs/OPinit/x/ophost/testutil"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

func Test_RegisterMigrationInfo_Success(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Test successful registration
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		1,
		migrationInfo,
	)

	_, err := ms.RegisterMigrationInfo(ctx, msg)
	require.ErrorContains(t, err, "bridge not found")

	// create bridge
	_, err = ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		BatchInfo: ophosttypes.BatchInfo{
			Submitter: testutil.AddrsStr[0],
			ChainType: ophosttypes.BatchInfo_INITIA,
		},
	}))
	require.NoError(t, err)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Verify migration info was stored
	hasInfo, err := input.OPHostKeeper.HasMigrationInfo(ctx, 1, "test1")
	require.NoError(t, err)
	require.True(t, hasInfo)

	storedInfo, err := input.OPHostKeeper.GetMigrationInfo(ctx, 1, "test1")
	require.NoError(t, err)
	require.Equal(t, migrationInfo, storedInfo)
}

func Test_RegisterMigrationInfo_InvalidAuthority(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Test with invalid authority - use a valid address format but not the gov module address
	invalidAuthority := testutil.AddrsStr[0] // This is not the gov module address
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		invalidAuthority,
		1,
		migrationInfo,
	)

	_, err := ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func Test_RegisterMigrationInfo_DuplicateRegistration(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// First register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		1,
		migrationInfo,
	)

	// create bridge
	_, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		BatchInfo: ophosttypes.BatchInfo{
			Submitter: testutil.AddrsStr[0],
			ChainType: ophosttypes.BatchInfo_INITIA,
		},
	}))
	require.NoError(t, err)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Try to register the same migration info again
	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration info already exists")
}

func Test_RegisterMigrationInfo_MoveEscrowedFunds(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Fund the bridge account
	bridgeAddr := ophosttypes.BridgeAddress(createRes.BridgeId)
	input.Faucet.Fund(ctx, bridgeAddr, sdk.NewCoin("test1", math.NewInt(1000)))

	// Check initial balance
	initialBalance := input.BankKeeper.GetBalance(ctx, bridgeAddr, "test1")
	require.Equal(t, math.NewInt(1000), initialBalance.Amount)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Verify funds were moved to IBC escrow account
	transferEscrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	finalBalance := input.BankKeeper.GetBalance(ctx, bridgeAddr, "test1")
	escrowBalance := input.BankKeeper.GetBalance(ctx, transferEscrowAddress, "test1")

	require.Equal(t, math.NewInt(0), finalBalance.Amount)
	require.Equal(t, math.NewInt(1000), escrowBalance.Amount)
}

func Test_HandleMigratedTokenDeposit_NonMigratedToken(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	input.Faucet.Fund(ctx, testutil.Addrs[0], sdk.NewCoin("non-migrated-token", math.NewInt(100)))

	// Test non-migrated token deposit
	depositMsg := ophosttypes.NewMsgInitiateTokenDeposit(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		testutil.AddrsStr[1], // to
		sdk.NewCoin("non-migrated-token", math.NewInt(100)),
		[]byte("test data"),
	)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	_, err = ms.InitiateTokenDeposit(ctx, depositMsg)
	require.NoError(t, err)
}

func Test_MigrationInfo_CRUD_Operations(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// Test setting migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     1,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	err := input.OPHostKeeper.SetMigrationInfo(ctx, migrationInfo)
	require.NoError(t, err)

	// Test getting migration info
	retrievedInfo, err := input.OPHostKeeper.GetMigrationInfo(ctx, 1, "test1")
	require.NoError(t, err)
	require.Equal(t, migrationInfo, retrievedInfo)

	// Test checking if migration info exists
	exists, err := input.OPHostKeeper.HasMigrationInfo(ctx, 1, "test1")
	require.NoError(t, err)
	require.True(t, exists)

	// Test checking non-existent migration info
	exists, err = input.OPHostKeeper.HasMigrationInfo(ctx, 1, "non-existent")
	require.NoError(t, err)
	require.False(t, exists)
}

func Test_HandleMigratedTokenDeposit_IBCTransferFailure(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test migrated token deposit
	depositMsg := ophosttypes.NewMsgInitiateTokenDeposit(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		testutil.AddrsStr[1], // to
		sdk.NewCoin("test1", math.NewInt(100)),
		[]byte("test data"),
	)

	input.MockRouter.SetShouldFail(true)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	_, err = ms.InitiateTokenDeposit(ctx, depositMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), sdkerrors.ErrInvalidRequest.Error())
}

func Test_HandleMigratedTokenDeposit_IBCTransferSuccess(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test migrated token deposit
	depositMsg := ophosttypes.NewMsgInitiateTokenDeposit(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		testutil.AddrsStr[1], // to
		sdk.NewCoin("test1", math.NewInt(100)),
		[]byte("test data"),
	)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	_, err = ms.InitiateTokenDeposit(ctx, depositMsg)
	require.NoError(t, err)

	// Verify the transfer event was emitted
	events := ctx.EventManager().Events()
	require.Len(t, events, 1)
	require.Equal(t, "ibc_transfer", events[0].Type)
}

func Test_HandleMigratedTokenWithdrawal_Direct(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the IBC escrow account
	transferEscrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	input.Faucet.Fund(ctx, transferEscrowAddress, sdk.NewCoin("test1", math.NewInt(1000)))

	// Check initial balances
	initialEscrowBalance := input.BankKeeper.GetBalance(ctx, transferEscrowAddress, "test1")
	initialReceiverBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], "test1")
	require.Equal(t, math.NewInt(1000), initialEscrowBalance.Amount)
	require.Equal(t, math.NewInt(0), initialReceiverBalance.Amount)

	// Test migrated token withdrawal directly
	withdrawalMsg := ophosttypes.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		1,                    // outputIndex
		1,                    // sequence
		[][]byte{},           // withdrawalProofs
		testutil.AddrsStr[0], // from
		testutil.AddrsStr[1], // to
		sdk.NewCoin("test1", math.NewInt(500)),
		[]byte{1},        // version
		make([]byte, 32), // storageRoot
		make([]byte, 32), // lastBlockHash
	)

	handled, err := input.OPHostKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.NoError(t, err)
	require.True(t, handled)

	// Check final balances
	finalEscrowBalance := input.BankKeeper.GetBalance(ctx, transferEscrowAddress, "test1")
	finalReceiverBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], "test1")
	require.Equal(t, math.NewInt(500), finalEscrowBalance.Amount) // 1000 - 500
	require.Equal(t, math.NewInt(500), finalReceiverBalance.Amount)
}

func Test_HandleMigratedTokenWithdrawal_NonMigratedToken_Direct(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Fund the bridge account (not IBC escrow)
	bridgeAddr := ophosttypes.BridgeAddress(createRes.BridgeId)
	input.Faucet.Fund(ctx, bridgeAddr, sdk.NewCoin("non-migrated-token", math.NewInt(1000)))

	// Check initial balances
	initialBridgeBalance := input.BankKeeper.GetBalance(ctx, bridgeAddr, "non-migrated-token")
	initialReceiverBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], "non-migrated-token")
	require.Equal(t, math.NewInt(1000), initialBridgeBalance.Amount)
	require.Equal(t, math.NewInt(0), initialReceiverBalance.Amount)

	// Test non-migrated token withdrawal directly
	withdrawalMsg := ophosttypes.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		1,                    // outputIndex
		1,                    // sequence
		[][]byte{},           // withdrawalProofs
		testutil.AddrsStr[0], // from
		testutil.AddrsStr[1], // to
		sdk.NewCoin("non-migrated-token", math.NewInt(500)),
		[]byte{1},        // version
		make([]byte, 32), // storageRoot
		make([]byte, 32), // lastBlockHash
	)

	handled, err := input.OPHostKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.NoError(t, err)
	require.False(t, handled) // Should not handle non-migrated tokens

	// Balances should remain unchanged since it wasn't handled
	finalBridgeBalance := input.BankKeeper.GetBalance(ctx, bridgeAddr, "non-migrated-token")
	finalReceiverBalance := input.BankKeeper.GetBalance(ctx, testutil.Addrs[1], "non-migrated-token")
	require.Equal(t, math.NewInt(1000), finalBridgeBalance.Amount) // unchanged
	require.Equal(t, math.NewInt(0), finalReceiverBalance.Amount)  // unchanged
}

func Test_HandleMigratedTokenWithdrawal_InvalidReceiverAddress_Direct(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the IBC escrow account
	transferEscrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	input.Faucet.Fund(ctx, transferEscrowAddress, sdk.NewCoin("test1", math.NewInt(1000)))

	// Test withdrawal with invalid receiver address
	withdrawalMsg := ophosttypes.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		1,                    // outputIndex
		1,                    // sequence
		[][]byte{},           // withdrawalProofs
		testutil.AddrsStr[0], // from
		"invalid-address",    // invalid address
		sdk.NewCoin("test1", math.NewInt(500)),
		[]byte{1},        // version
		make([]byte, 32), // storageRoot
		make([]byte, 32), // lastBlockHash
	)

	_, err = input.OPHostKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed")
}

func Test_HandleMigratedTokenWithdrawal_InsufficientBalance_Direct(t *testing.T) {
	ctx, input := testutil.CreateTestInput(t, false)

	// Create a bridge first
	bridgeConfig := ophosttypes.BridgeConfig{
		Challenger:            testutil.AddrsStr[0],
		Proposer:              testutil.AddrsStr[0],
		SubmissionInterval:    time.Second * 10,
		FinalizationPeriod:    time.Second * 60,
		SubmissionStartHeight: 1,
		Metadata:              []byte{1, 2, 3},
		BatchInfo:             ophosttypes.BatchInfo{Submitter: testutil.AddrsStr[0], ChainType: ophosttypes.BatchInfo_INITIA},
	}

	ms := keeper.NewMsgServerImpl(input.OPHostKeeper)
	createRes, err := ms.CreateBridge(ctx, ophosttypes.NewMsgCreateBridge(testutil.AddrsStr[0], bridgeConfig))
	require.NoError(t, err)

	// Register migration info
	migrationInfo := ophosttypes.MigrationInfo{
		BridgeId:     createRes.BridgeId,
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
		L1Denom:      "test1",
	}

	msg := ophosttypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		createRes.BridgeId,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the IBC escrow account with insufficient balance
	transferEscrowAddress := transfertypes.GetEscrowAddress("transfer", "channel-0")
	input.Faucet.Fund(ctx, transferEscrowAddress, sdk.NewCoin("test1", math.NewInt(100)))

	// Test withdrawal with amount larger than available balance
	withdrawalMsg := ophosttypes.NewMsgFinalizeTokenWithdrawal(
		testutil.AddrsStr[0],
		createRes.BridgeId,
		1,                                      // outputIndex
		1,                                      // sequence
		[][]byte{},                             // withdrawalProofs
		testutil.AddrsStr[0],                   // from
		testutil.AddrsStr[1],                   // to
		sdk.NewCoin("test1", math.NewInt(500)), // more than available 100
		[]byte{1},                              // version
		make([]byte, 32),                       // storageRoot
		make([]byte, 32),                       // lastBlockHash
	)

	_, err = input.OPHostKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient funds")
}
