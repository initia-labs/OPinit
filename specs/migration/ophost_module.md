# OPHost Asset Migration

## Overview

The OPHost module provides **migration handling** functionality that enables switching from OP Bridge to IBC Bridge while maintaining seamless user experience. It handles L1 → L2 token migration via IBC transfer and L2 → L1 withdrawal from IBC escrow accounts, ensuring users continue using existing bridge messages while the underlying infrastructure switches to IBC. With the IBC integration, users can now also use `MsgTransfer` as a new interface to initiate token deposits, providing more flexibility in how tokens are transferred between chains.

## Core Functions

### 1. Migration Info Management

#### `SetMigrationInfo`

- **Purpose**: Registers migration configuration for bridge replacement
- **Parameters**: `MigrationInfo` struct containing bridge ID, L1 denom, IBC channel ID, and port ID
- **Storage**: Maps bridge ID + L1 denom to migration configuration

#### `GetMigrationInfo`

- **Purpose**: Retrieves migration configuration for a specific bridge and L1 denom
- **Parameters**: Bridge ID and L1 denom
- **Returns**: `MigrationInfo` struct or error if not found

#### `HasMigrationInfo`

- **Purpose**: Checks if migration info exists for a given bridge and L1 denom
- **Parameters**: Bridge ID and L1 denom
- **Returns**: Boolean indicating existence

#### `IterateMigrationInfos`

- **Purpose**: Iterates over all registered migration infos
- **Usage**: Callback-based iteration for bulk operations

### 2. Bridge Handling

#### `HandleMigratedTokenDeposit`

- **Purpose**: Processes migrated token deposits via IBC transfer
- **Process**:
  1. Checks if L1 denom has migration info registered
  2. Creates IBC transfer message with migration parameters
  3. Routes transfer through message router
  4. Emits events for tracking
- **Returns**: Boolean indicating if handled, plus any error
- **Integration**: Seamlessly integrates with existing `MsgInitiateTokenDeposit` workflow

#### `HandleMigratedTokenWithdrawal`

- **Purpose**: Processes in-flight withdrawal requests that were initiated before migration was registered
- **Process**:
  1. Checks if L1 denom has migration info registered
  2. Retrieves IBC escrow address from migration info
  3. Transfers tokens from IBC escrow to receiver address
  4. Returns `true` if handled, `false` if not migrated (falls back to bridge withdrawal)
- **Returns**: Boolean indicating if handled, plus any error
- **Integration**: Integrates with `MsgFinalizeTokenWithdrawal` workflow for in-flight requests
- **Fallback**: If not migrated, normal bridge withdrawal logic takes over
- **Use Case**: Handles withdrawal requests that were initiated before migration registration

## Integration

### 1. User Experience Preservation

- **Same Messages**: Users continue using `MsgInitiateTokenDeposit` and `MsgFinalizeTokenWithdrawal`
- **Same Tokens**: Users work with L1 tokens (ex: INIT)
- **Hidden Complexity**: Bridge replacement is transparent to users

### 2. IBC Transfer Integration

- **Message Router**: Routes IBC transfer messages via `msgRouter`
- **Transfer Creation**: Creates `MsgTransfer` with migration parameters
- **Event Emission**: Emits transfer events for tracking

### 3. IBC Escrow Integration

- **Escrow Withdrawal**: Handles withdrawals from IBC transfer escrow accounts
- **Address Resolution**: Uses migration info to determine correct escrow address
- **Balance Verification**: Ensures sufficient funds in escrow before withdrawal
- **Error Handling**: Proper error propagation for insufficient funds or invalid addresses

### 4. Bridge Hook Support

- **Hook Preservation**: Maintains OP Bridge hook functionality during bridge replacement
- **Memo Encoding**: Encodes bridge hook data in IBC transfer message's memo field
- **Hook Data Format**: Uses `MigratedTokenDepositMemo` structure to preserve hook information
- **Cross-Chain Execution**: Hook data travels via IBC and gets executed on OPChild side

## Migrated Token Bridge Flow

### Hook Preservation Strategy

The migration system preserves the existing OP Bridge hook functionality by encoding hook data in IBC transfer messages and executing it on the receiving side.

### Complete Hook Flow

```plaintext
1. User calls MsgInitiateTokenDeposit with hook data
2. OPHost encodes hook data as MigratedTokenDepositMemo
3. Hook data is embedded in IBC transfer message's memo field
4. IBC transfer sends tokens + hook data to L2
5. OPChild receives IBC tokens and decodes hook data
6. OPChild executes the hook using its execute hook function
7. User gets OP tokens + hook execution results
```

### L2 → L1 Flow (In-Flight Withdrawal Handling)

```plaintext
1. User initiated withdrawal before migration was registered
2. Migration gets registered (tokens moved to IBC escrow)
3. User calls MsgFinalizeTokenWithdrawal
4. OPHost checks if token has migration info
5. If migrated: Transfer from IBC escrow to receiver
6. If not migrated: Fall back to bridge account withdrawal
7. User receives L1 tokens (same as before)
```

### Hook Data Structure

- **Memo Format**: JSON-encoded `MigratedTokenDepositMemo`
- **Data Preservation**: All original hook information is maintained
- **Cross-Chain Transfer**: Hook data travels securely via IBC protocol

### Benefits of This Approach

- **Feature Preservation**: Users keep all existing hook functionality
- **Seamless Migration**: No changes needed to hook implementation
- **IBC Integration**: Leverages IBC's secure cross-chain communication
- **Execution Consistency**: Hooks execute the same way on both sides
- **Withdrawal Support**: Complete L2→L1 flow with IBC escrow integration
