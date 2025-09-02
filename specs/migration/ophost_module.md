# OPHost Asset Migration

## Overview

The OPHost module provides **migration handling** functionality that enables switching from OP Bridge to IBC Bridge while maintaining seamless user experience. It handles L1 → L2 token migration via IBC transfer, ensuring users continue using `MsgInitiateTokenDeposit` while the underlying infrastructure switches to IBC. With the IBC integration, users can now also use `MsgTransfer` as a new interface to initiate token deposits, providing more flexibility in how tokens are transferred between chains.

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

## Integration

### 1. User Experience Preservation

- **Same Messages**: Users continue using `MsgInitiateTokenDeposit`
- **Same Tokens**: Users work with L1 tokens (ex: INIT)
- **Hidden Complexity**: Bridge replacement is transparent to users

### 2. IBC Transfer Integration

- **Message Router**: Routes IBC transfer messages via `msgRouter`
- **Transfer Creation**: Creates `MsgTransfer` with migration parameters
- **Event Emission**: Emits transfer events for tracking

### 3. Bridge Hook Support

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

### Hook Data Structure

- **Memo Format**: JSON-encoded `MigratedTokenDepositMemo`
- **Data Preservation**: All original hook information is maintained
- **Cross-Chain Transfer**: Hook data travels securely via IBC protocol

### Benefits of This Approach

- **Feature Preservation**: Users keep all existing hook functionality
- **Seamless Migration**: No changes needed to hook implementation
- **IBC Integration**: Leverages IBC's secure cross-chain communication
- **Execution Consistency**: Hooks execute the same way on both sides
