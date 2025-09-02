# OPChild Asset Migration

## Overview

The OPChild module provides comprehensive **bridge replacement** functionality that enables switching from OP Bridge to IBC Bridge while maintaining seamless user experience. It allows users to continue using existing OP INIT tokens (created via OP Bridge) while the underlying infrastructure switches to IBC, ensuring users receive OP INIT tokens instead of IBC INIT tokens with minimal side effects.

## Core Migration Functions

### 1. Migration Info Management

#### `SetMigrationInfo`

- **Purpose**: Registers migration configuration for a specific denom
- **Parameters**: `MigrationInfo` struct containing denom, IBC channel ID, and port ID
- **Validation**: Ensures valid channel identifier, port identifier, and denom format

#### `GetMigrationInfo`

- **Purpose**: Retrieves migration configuration for a specific denom
- **Returns**: `MigrationInfo` struct or error if not found

#### `HasMigrationInfo`

- **Purpose**: Checks if migration info exists for a given denom
- **Returns**: Boolean indicating existence

#### `IterateMigrationInfos`

- **Purpose**: Iterates over all registered migration infos
- **Usage**: Callback-based iteration for bulk operations

### 2. IBC to L2 Denom Mapping

#### `SetIBCToL2DenomMap`

- **Purpose**: Maps IBC denoms to their corresponding L2 denoms
- **Usage**: Essential for automatic IBC deposit handling (IBC → L2)

#### `GetIBCToL2DenomMap`

- **Purpose**: Retrieves L2 denom for a given IBC denom
- **Returns**: L2 denom string or error

#### `HasIBCToL2DenomMap`

- **Purpose**: Checks if IBC denom mapping exists
- **Returns**: Boolean indicating existence

### 3. Token Migration Operations

#### `MigrateToken` (Forward Migration: L2 → IBC)

- **Purpose**: Enable explicit token migration from L2 OP tokens to IBC tokens for cross-chain transfers
- **Process**:
  1. Validates positive amount and denom match
  2. Transfers tokens to module account
  3. Burns L2 tokens from module
  4. Mints IBC tokens to module
  5. Sends IBC tokens to sender
- **Returns**: Minted IBC coin

#### `HandleMigratedTokenDeposit` (IBC Deposit Handling: IBC → L2)

- **Purpose**: **Core logic to convert IBC tokens to OP tokens** - processes incoming IBC tokens and converts them to L2 tokens
- **Process**:
  1. Validates positive amount
  2. Retrieves L2 denom mapping
  3. Transfers IBC tokens to module
  4. Burns IBC tokens
  5. Mints L2 tokens
  6. Sends L2 tokens to receiver
  7. **Optional**: Executes bridge hooks if present in memo
- **Returns**: Minted L2 coin
- **Note**: This is automatic handling of IBC deposits (IBC → L2 conversion)

#### `HandleMigratedTokenWithdrawal`

- **Purpose**: Handles withdrawal requests that trigger token migration
- **Process**:
  1. Checks if denom has migration info
  2. Migrates L2 tokens to IBC tokens
  3. Creates IBC transfer message
  4. Routes transfer through message router
  5. Emits events for tracking
- **Returns**: Boolean indicating if handled, plus any error

## Message Types

### `MsgRegisterMigrationInfo`

- **Purpose**: Registers new migration configuration
- **Authority**: Module authority required
- **Validation**: Comprehensive validation of all fields

### `MsgMigrateToken`

- **Purpose**: Enable explicit token migration from L2 OP tokens to IBC tokens for cross-chain transfers
- **Parameters**: Sender address, amount, migration info
- **Validation**: Amount must be positive, denom must match

### `MsgInitiateTokenWithdrawal`

- **Purpose**: Initiates withdrawal with automatic migration
- **Integration**: Automatically triggers migration if configured
- **Fallback**: Passes through to standard withdrawal if no migration
