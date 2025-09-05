package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/initia-labs/OPinit/x/opchild/keeper"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
)

// Test_RegisterMigrationInfo_Success tests successful registration of migration info
func Test_RegisterMigrationInfo_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Test successful registration
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Verify migration info was stored using keeper functions
	hasInfo, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.True(t, hasInfo)

	storedInfo, err := input.OPChildKeeper.GetMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.Equal(t, migrationInfo, storedInfo)
}

// Test_RegisterMigrationInfo_InvalidAuthority tests registration with invalid authority
func Test_RegisterMigrationInfo_InvalidAuthority(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Test with invalid authority - use a valid address format but not the opchild module address
	invalidAuthority := addrsStr[0] // This is not the opchild module address
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		invalidAuthority,
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

// Test_RegisterMigrationInfo_DuplicateRegistration tests duplicate registration prevention
func Test_RegisterMigrationInfo_DuplicateRegistration(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// First register migration info
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Try to register the same migration info again
	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration info already registered")
}

// Test_RegisterMigrationInfo_InvalidParameters tests various invalid parameter scenarios
func Test_RegisterMigrationInfo_InvalidParameters(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Test with empty IBC channel ID
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err := ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)

	// Test with empty IBC port ID
	migrationInfo = opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "",
	}

	msg = opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)

	// Test with empty denom
	migrationInfo = opchildtypes.MigrationInfo{
		Denom:        "",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg = opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.Error(t, err)
}

// Test_MigrationInfo_CRUD_Operations tests basic CRUD operations for migration info
func Test_MigrationInfo_CRUD_Operations(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Test setting migration info
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test getting migration info
	storedInfo, err := input.OPChildKeeper.GetMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.Equal(t, migrationInfo, storedInfo)

	// Test checking if migration info exists
	hasInfo, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.True(t, hasInfo)

	// Test checking non-existent migration info
	exists, err := input.OPChildKeeper.HasMigrationInfo(ctx, "non-existent")
	require.NoError(t, err)
	require.False(t, exists)
}

// Test_MigrationInfo_MultipleDenoms tests migration info across multiple denoms
func Test_MigrationInfo_MultipleDenoms(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pairs first (L1 tokens)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)
	err = input.OPChildKeeper.DenomPairs.Set(ctx, "test2", "test2")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Register migration info for first denom
	migrationInfo1 := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg1 := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo1,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg1)
	require.NoError(t, err)

	// Register migration info for second denom
	migrationInfo2 := opchildtypes.MigrationInfo{
		Denom:        "test2",
		IbcChannelId: "channel-1",
		IbcPortId:    "transfer",
	}

	msg2 := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo2,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg2)
	require.NoError(t, err)

	// Verify both migration infos are stored correctly
	hasInfo1, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.True(t, hasInfo1)

	hasInfo2, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test2")
	require.NoError(t, err)
	require.True(t, hasInfo2)

	// Verify they don't interfere with each other
	hasInfo1WrongDenom, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test2")
	require.NoError(t, err)
	require.True(t, hasInfo1WrongDenom) // test2 should exist

	hasInfo2WrongDenom, err := input.OPChildKeeper.HasMigrationInfo(ctx, "test1")
	require.NoError(t, err)
	require.True(t, hasInfo2WrongDenom) // test1 should exist
}

// Test_MigrationInfo_Iteration tests iteration over migration infos
func Test_MigrationInfo_Iteration(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pairs first (L1 tokens)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)
	err = input.OPChildKeeper.DenomPairs.Set(ctx, "test2", "test2")
	require.NoError(t, err)
	err = input.OPChildKeeper.DenomPairs.Set(ctx, "test3", "test3")
	require.NoError(t, err)

	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)

	// Register multiple migration infos
	migrationInfos := []opchildtypes.MigrationInfo{
		{
			Denom:        "test1",
			IbcChannelId: "channel-0",
			IbcPortId:    "transfer",
		},
		{
			Denom:        "test2",
			IbcChannelId: "channel-1",
			IbcPortId:    "transfer",
		},
		{
			Denom:        "test3",
			IbcChannelId: "channel-2",
			IbcPortId:    "transfer",
		},
	}

	for _, info := range migrationInfos {
		msg := opchildtypes.NewMsgRegisterMigrationInfo(
			authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
			info,
		)
		_, err := ms.RegisterMigrationInfo(ctx, msg)
		require.NoError(t, err)
	}

	// Test iteration
	var collectedInfos []opchildtypes.MigrationInfo
	err = input.OPChildKeeper.IterateMigrationInfos(ctx, func(denom string, migrationInfo opchildtypes.MigrationInfo) (stop bool, err error) {
		collectedInfos = append(collectedInfos, migrationInfo)
		return false, nil
	})
	require.NoError(t, err)
	require.Len(t, collectedInfos, 3)

	// Verify all infos were collected
	require.Contains(t, collectedInfos, migrationInfos[0])
	require.Contains(t, collectedInfos, migrationInfos[1])
	require.Contains(t, collectedInfos, migrationInfos[2])
}

// Test_MigrationInfo_QueryNonExistent tests querying non-existent migration info
func Test_MigrationInfo_QueryNonExistent(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Test querying non-existent migration info
	_, err := input.OPChildKeeper.GetMigrationInfo(ctx, "non-existent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Test checking non-existent migration info
	hasInfo, err := input.OPChildKeeper.HasMigrationInfo(ctx, "non-existent")
	require.NoError(t, err)
	require.False(t, hasInfo)
}

// Test_IBCToL2DenomMap_CRUD tests IBC to L2 denom mapping functionality
func Test_IBCToL2DenomMap_CRUD(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Test setting IBC to L2 denom map
	ibcDenom := "ibc/1234567890ABCDEF"
	l2Denom := "test1"
	err := input.OPChildKeeper.SetIBCToL2DenomMap(ctx, ibcDenom, l2Denom)
	require.NoError(t, err)

	// Test getting IBC to L2 denom map
	retrievedL2Denom, err := input.OPChildKeeper.GetIBCToL2DenomMap(ctx, ibcDenom)
	require.NoError(t, err)
	require.Equal(t, l2Denom, retrievedL2Denom)

	// Test checking if IBC to L2 denom map exists
	hasMap, err := input.OPChildKeeper.HasIBCToL2DenomMap(ctx, ibcDenom)
	require.NoError(t, err)
	require.True(t, hasMap)

	// Test checking non-existent IBC to L2 denom map
	hasNonExistentMap, err := input.OPChildKeeper.HasIBCToL2DenomMap(ctx, "non-existent")
	require.NoError(t, err)
	require.False(t, hasNonExistentMap)
}

// Test_MigrateToken_Success tests successful token migration
func Test_MigrateToken_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the sender account
	sender := addrs[0]
	amount := sdk.NewCoin("test1", math.NewInt(100))
	input.Faucet.Fund(ctx, sender, sdk.NewCoins(amount)...)

	// Test token migration
	_, err = ms.MigrateToken(ctx, &opchildtypes.MsgMigrateToken{
		Sender: sender.String(),
		Amount: amount,
	})
	require.NoError(t, err)

	// Verify the IBC coin was created correctly
	expectedIBCDenom := transfertypes.GetTransferCoin(migrationInfo.IbcPortId, migrationInfo.IbcChannelId, "test1", amount.Amount)
	ibcCoin := input.BankKeeper.GetBalance(ctx, sender, expectedIBCDenom.Denom)
	require.Equal(t, expectedIBCDenom, ibcCoin)

	// Verify sender balance is now 0 for the original token
	finalBalance := input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(0), finalBalance.Amount)

	// Verify sender now has the IBC token
	ibcBalance := input.BankKeeper.GetBalance(ctx, sender, ibcCoin.Denom)
	require.Equal(t, amount.Amount, ibcBalance.Amount)
}

// Test_MigrateToken_InvalidAmount tests token migration with invalid amount
func Test_MigrateToken_InvalidAmount(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test with zero amount
	zeroAmount := sdk.NewCoin("test1", math.NewInt(0))
	_, err = ms.MigrateToken(ctx, &opchildtypes.MsgMigrateToken{
		Sender: addrs[0].String(),
		Amount: zeroAmount,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")

	// Test with negative amount - create coin directly to avoid panic
	negativeAmount := sdk.Coin{
		Denom:  "test1",
		Amount: math.NewInt(-100),
	}
	_, err = ms.MigrateToken(ctx, &opchildtypes.MsgMigrateToken{
		Sender: addrs[0].String(),
		Amount: negativeAmount,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid amount")
}

// Test_MigrateToken_NotFound tests token migration with not found migration info
func Test_MigrateToken_NotFound(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test with mismatched denom
	wrongDenomAmount := sdk.NewCoin("test2", math.NewInt(100))
	_, err = ms.MigrateToken(ctx, &opchildtypes.MsgMigrateToken{
		Sender: addrs[0].String(),
		Amount: wrongDenomAmount,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "migration info not found")
}

// Test_MigrateToken_InsufficientBalance tests token migration with insufficient balance
func Test_MigrateToken_InsufficientBalance(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test with amount larger than balance
	largeAmount := sdk.NewCoin("test1", math.NewInt(1000))
	_, err = ms.MigrateToken(ctx, &opchildtypes.MsgMigrateToken{
		Sender: addrs[0].String(),
		Amount: largeAmount,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient funds")
}

// Test_HandleMigratedTokenWithdrawal_Success tests successful handling of migrated token withdrawal
func Test_HandleMigratedTokenWithdrawal_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the sender account
	sender := addrs[0]
	amount := sdk.NewCoin("test1", math.NewInt(100))
	input.Faucet.Fund(ctx, sender, sdk.NewCoins(amount)...)

	// Create withdrawal message
	withdrawalMsg := opchildtypes.NewMsgInitiateTokenWithdrawal(
		sender.String(),
		addrs[1].String(), // to
		amount,
	)

	// Test handling migrated token withdrawal
	handled, err := input.OPChildKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.NoError(t, err)
	require.True(t, handled)

	// Get the actual IBC denom that was created
	baseDenom, err := input.OPChildKeeper.GetBaseDenom(ctx, "test1")
	require.NoError(t, err)
	expectedIBCDenom := transfertypes.GetTransferCoin(migrationInfo.IbcPortId, migrationInfo.IbcChannelId, baseDenom, amount.Amount)

	// Verify the token was migrated and IBC token was created
	ibcBalance := input.BankKeeper.GetBalance(ctx, sender, expectedIBCDenom.Denom)
	require.Equal(t, amount.Amount, ibcBalance.Amount)

	// Verify original token balance is 0
	originalBalance := input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(0), originalBalance.Amount)

	// check handled messages
	handledMsgs := input.MockRouter.GetHandledMsgs()
	require.Len(t, handledMsgs, 1)
	handledMsgs[0].TimeoutHeight = clienttypes.NewHeight(0, 0)
	handledMsgs[0].TimeoutTimestamp = 0
	require.Equal(t, handledMsgs[0], &transfertypes.MsgTransfer{
		Sender:        sender.String(),
		Receiver:      addrs[1].String(),
		Token:         sdk.NewCoin(expectedIBCDenom.Denom, amount.Amount),
		SourcePort:    migrationInfo.IbcPortId,
		SourceChannel: migrationInfo.IbcChannelId,
		Memo:          "forwarded from opchild module",
	})
}

// Test_HandleMigratedTokenWithdrawal_NonMigratedToken tests handling withdrawal of non-migrated token
func Test_HandleMigratedTokenWithdrawal_NonMigratedToken(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Fund the sender account with a non-migrated token
	sender := addrs[0]
	amount := sdk.NewCoin("non-migrated-token", math.NewInt(100))
	input.Faucet.Fund(ctx, sender, sdk.NewCoins(amount)...)

	// Create withdrawal message
	withdrawalMsg := opchildtypes.NewMsgInitiateTokenWithdrawal(
		sender.String(),
		addrs[1].String(), // to
		amount,
	)

	// Test handling non-migrated token withdrawal
	handled, err := input.OPChildKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.NoError(t, err)
	require.False(t, handled) // Should not be handled since it's not a migrated token

	// Verify the token balance remains unchanged
	balance := input.BankKeeper.GetBalance(ctx, sender, "non-migrated-token")
	require.Equal(t, amount.Amount, balance.Amount)
}

// Test_HandleMigratedTokenWithdrawal_InvalidSender tests handling withdrawal with invalid sender
func Test_HandleMigratedTokenWithdrawal_InvalidSender(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Create withdrawal message with invalid sender
	withdrawalMsg := opchildtypes.NewMsgInitiateTokenWithdrawal(
		"invalid-address",
		addrs[1].String(), // to
		sdk.NewCoin("test1", math.NewInt(100)),
	)

	// Test handling withdrawal with invalid sender
	_, err = input.OPChildKeeper.HandleMigratedTokenWithdrawal(ctx, withdrawalMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decoding bech32 failed")
}

// Test_HandleMigratedTokenDeposit_Success tests successful handling of migrated token deposit
func Test_HandleMigratedTokenDeposit_Success(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the sender account with IBC token
	sender := addrs[0]
	ibcCoin := sdk.NewCoin("ibc/1234567890ABCDEF", math.NewInt(100))

	// Set up IBC to L2 denom mapping
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/1234567890ABCDEF", "test1")
	require.NoError(t, err)

	// Fund the sender account with IBC token
	input.Faucet.Fund(ctx, sender, ibcCoin)

	// Test migrated token deposit
	_, err = ms.HandleMigratedTokenDeposit(ctx, sender, ibcCoin, "")
	require.NoError(t, err)

	// Verify the IBC token was burned and L2 token was created
	ibcBalance := input.BankKeeper.GetBalance(ctx, sender, "ibc/1234567890ABCDEF")
	require.Equal(t, math.NewInt(0), ibcBalance.Amount)

	// Verify sender now has the L2 token
	l2Balance := input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(100), l2Balance.Amount)

	//////////////////////////////////////////////

	// create hook data
	priv, _, addr := keyPubAddr()

	// Fund the sender account with IBC token
	input.Faucet.Fund(ctx, addr, ibcCoin)
	halfAmount := ibcCoin.Amount.QuoRaw(2)

	acc := input.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}
	from, _ := input.AccountKeeper.AddressCodec().BytesToString(addr)
	to, _ := input.AccountKeeper.AddressCodec().BytesToString(addrs[2])

	signedTxBz, err := input.EncodingConfig.TxConfig.TxEncoder()(generateTestTx(
		t, input,
		[]sdk.Msg{
			opchildtypes.NewMsgInitiateTokenWithdrawal(from, to, sdk.NewCoin("test1", halfAmount)),
		},
		privs, accNums, accSeqs, sdk.UnwrapSDKContext(ctx).ChainID(),
	))
	require.NoError(t, err)

	// Test migrated token deposit with hook
	_, err = ms.HandleMigratedTokenDeposit(ctx, addr, ibcCoin, fmt.Sprintf(`{"opinit": "%s"}`, base64.StdEncoding.EncodeToString(signedTxBz)))
	require.NoError(t, err)

	// check balance
	ibcBalance = input.BankKeeper.GetBalance(ctx, addr, "ibc/1234567890ABCDEF")
	require.Equal(t, math.ZeroInt(), ibcBalance.Amount)

	// halfAmount should be withdrawn from the account
	l2Balance = input.BankKeeper.GetBalance(ctx, addr, "test1")
	require.Equal(t, halfAmount, l2Balance.Amount)
}

// Test_HandleMigratedTokenDeposit_InvalidAmount tests handling of migrated token deposit with invalid amount
func Test_HandleMigratedTokenDeposit_InvalidAmount(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test with zero amount
	zeroAmount := sdk.NewCoin("ibc/1234567890ABCDEF", math.NewInt(0))

	_, err = ms.HandleMigratedTokenDeposit(ctx, addrs[0], zeroAmount, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "amount is not positive")

	// Test with negative amount - create coin directly to avoid panic
	negativeAmount := sdk.Coin{
		Denom:  "ibc/1234567890ABCDEF",
		Amount: math.NewInt(-100),
	}

	_, err = ms.HandleMigratedTokenDeposit(ctx, addrs[0], negativeAmount, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "amount is not positive")
}

// Test_HandleMigratedTokenDeposit_NonExistentIBCDenom tests handling of migrated token deposit with non-existent IBC denom
func Test_HandleMigratedTokenDeposit_NonExistentIBCDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Test with non-existent IBC denom
	nonExistentIBCAmount := sdk.NewCoin("ibc/NONEXISTENT", math.NewInt(100))
	_, err = ms.HandleMigratedTokenDeposit(ctx, addrs[0], nonExistentIBCAmount, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// Test_HandleMigratedTokenDeposit_InsufficientBalance tests handling of migrated token deposit with insufficient balance
func Test_HandleMigratedTokenDeposit_InsufficientBalance(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the sender account with a small amount of IBC token
	sender := addrs[0]
	ibcAmount := sdk.NewCoin("ibc/1234567890ABCDEF", math.NewInt(50))

	// Set up IBC to L2 denom mapping
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/1234567890ABCDEF", "test1")
	require.NoError(t, err)

	// Fund the sender account with IBC token
	input.Faucet.Fund(ctx, sender, ibcAmount)

	// Try to reverse migrate more than available
	largeAmount := sdk.NewCoin("ibc/1234567890ABCDEF", math.NewInt(100))
	_, err = ms.HandleMigratedTokenDeposit(ctx, sender, largeAmount, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient funds")
}

// Test_HandleMigratedTokenDeposit_CompleteFlow tests the complete flow: migrate then handle migrated token deposit
func Test_HandleMigratedTokenDeposit_CompleteFlow(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pair first (L1 token)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)

	// Register migration info first
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg)
	require.NoError(t, err)

	// Fund the sender account with L2 token
	sender := addrs[0]
	amount := sdk.NewCoin("test1", math.NewInt(100))
	input.Faucet.Fund(ctx, sender, sdk.NewCoins(amount)...)

	// Step 1: Migrate L2 token to IBC token
	ibcCoin, err := input.OPChildKeeper.MigrateToken(ctx, migrationInfo, sender, amount)
	require.NoError(t, err)

	// Verify L2 token is burned and IBC token is created
	l2Balance := input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(0), l2Balance.Amount)

	ibcBalance := input.BankKeeper.GetBalance(ctx, sender, ibcCoin.Denom)
	require.Equal(t, math.NewInt(100), ibcBalance.Amount)

	// Step 2: Reverse migrate IBC token back to L2 token
	_, err = ms.HandleMigratedTokenDeposit(ctx, sender, ibcCoin, "")
	require.NoError(t, err)

	// Verify IBC token is burned and L2 token is restored
	ibcBalance = input.BankKeeper.GetBalance(ctx, sender, ibcCoin.Denom)
	require.Equal(t, math.NewInt(0), ibcBalance.Amount)

	l2Balance = input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(100), l2Balance.Amount)
}

// Test_HandleMigratedTokenDeposit_MultipleDenoms tests handling of migrated token deposit across multiple denoms
func Test_HandleMigratedTokenDeposit_MultipleDenoms(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// Set up denom pairs first (L1 tokens)
	err := input.OPChildKeeper.DenomPairs.Set(ctx, "test1", "test1")
	require.NoError(t, err)
	err = input.OPChildKeeper.DenomPairs.Set(ctx, "test2", "test2")
	require.NoError(t, err)

	// Register migration info for first denom
	ms := keeper.NewMsgServerImpl(&input.OPChildKeeper)
	migrationInfo1 := opchildtypes.MigrationInfo{
		Denom:        "test1",
		IbcChannelId: "channel-0",
		IbcPortId:    "transfer",
	}

	msg1 := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo1,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg1)
	require.NoError(t, err)

	// Register migration info for second denom
	migrationInfo2 := opchildtypes.MigrationInfo{
		Denom:        "test2",
		IbcChannelId: "channel-1",
		IbcPortId:    "transfer",
	}

	msg2 := opchildtypes.NewMsgRegisterMigrationInfo(
		authtypes.NewModuleAddress(opchildtypes.ModuleName).String(),
		migrationInfo2,
	)

	_, err = ms.RegisterMigrationInfo(ctx, msg2)
	require.NoError(t, err)

	// Fund the sender account with both IBC tokens
	sender := addrs[0]
	ibcAmount1 := sdk.NewCoin("ibc/1234567890ABCDEF", math.NewInt(100))
	ibcAmount2 := sdk.NewCoin("ibc/FEDCBA0987654321", math.NewInt(200))

	// Set up IBC to L2 denom mappings
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/1234567890ABCDEF", "test1")
	require.NoError(t, err)
	err = input.OPChildKeeper.SetIBCToL2DenomMap(ctx, "ibc/FEDCBA0987654321", "test2")
	require.NoError(t, err)

	// Fund the sender account with IBC tokens
	input.Faucet.Fund(ctx, sender, sdk.NewCoins(ibcAmount1, ibcAmount2)...)

	// Reverse migrate first denom
	_, err = ms.HandleMigratedTokenDeposit(ctx, sender, ibcAmount1, "")
	require.NoError(t, err)

	// Verify first denom is reverse migrated
	l2Balance1 := input.BankKeeper.GetBalance(ctx, sender, "test1")
	require.Equal(t, math.NewInt(100), l2Balance1.Amount)

	ibcBalance1 := input.BankKeeper.GetBalance(ctx, sender, "ibc/1234567890ABCDEF")
	require.Equal(t, math.NewInt(0), ibcBalance1.Amount)

	// Reverse migrate second denom
	_, err = ms.HandleMigratedTokenDeposit(ctx, sender, ibcAmount2, "")
	require.NoError(t, err)

	// Verify second denom is reverse migrated
	l2Balance2 := input.BankKeeper.GetBalance(ctx, sender, "test2")
	require.Equal(t, math.NewInt(200), l2Balance2.Amount)

	ibcBalance2 := input.BankKeeper.GetBalance(ctx, sender, "ibc/FEDCBA0987654321")
	require.Equal(t, math.NewInt(0), ibcBalance2.Amount)

	// Verify they don't interfere with each other
	require.Equal(t, math.NewInt(100), l2Balance1.Amount)
	require.Equal(t, math.NewInt(200), l2Balance2.Amount)
}
