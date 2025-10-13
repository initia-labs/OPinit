# Bridge Replacement System Documentation

## Overview

The OPinit system implements a **bridge replacement mechanism** that switches from OP Bridge to IBC Bridge while preserving user experience. Users continue using existing `MsgInitiateTokenDeposit` and `MsgInitiateTokenWithdrawal` messages, but receive OP tokens instead of IBC tokens. The system automatically handles IBC → L2 conversion via middleware and supports bridge hook preservation through IBC transfer memos.

**Additionally, users can use normal IBC bridge functionality directly** for standard cross-chain transfers without the bridge replacement features.

## System Architecture

### Core Goal

Replace OP Bridge with IBC Bridge infrastructure while maintaining backward compatibility for users.

### Key Benefits

- **User Experience Preservation**: Users keep using existing messages and receive OP tokens
- **Infrastructure Modernization**: Switch to IBC Bridge for better interoperability
- **Feature Preservation**: Bridge hooks continue to work via IBC transfer memos
- **Seamless Migration**: No user-facing changes required
- **Dual Bridge Support**: Both bridge replacement and normal IBC bridge functionality available
- **In-Flight Request Handling**: Supports withdrawal requests initiated before migration

### Bridge Replacement Flows

#### L1 → L2 Flow (Two Options)

- Option A: Bridge Replacement (User Experience Preserved)

   ```plaintext
   User calls MsgInitiateTokenDeposit → Receives OP tokens on L2
   ```

- Option B: Normal IBC Bridge

   ```plaintext
   User calls MsgTransfer → IBC transfer → L2 receives IBC tokens → (opchild) convert IBC token to OP token on L2
   ```

#### L2 → L1 Flow (Two Options)

- Option A: Bridge Replacement (User Experience Preserved)

   ```plaintext
   User calls MsgInitiateTokenWithdrawal → Receives L1 tokens
   ```

- Option B: Explicit Migration

   ```plaintext
   User calls MsgMigrateToken → Gets IBC tokens → Manual IBC transfer
   ```

- Failure handling (applies to both options)

   If the outbound IBC transfer fails or times out, the middleware detects the refund packet and converts the returned IBC vouchers back into OP tokens for the user. For the explicit migration path, the user can re-run `MsgMigrateToken` to obtain IBC vouchers again before retrying.

#### Bridge Hook Preservation

```plaintext
1. User calls MsgInitiateTokenDeposit with hook data
2. OPHost encodes hook data as MigratedTokenDepositMemo
3. Hook data travels via IBC transfer memo
4. OPChild receives IBC tokens and decodes hook data
5. OPChild executes hook using existing hook function
6. User gets OP tokens + hook execution results
```

#### In-Flight Withdrawal Handling

```plaintext
1. User initiated withdrawal before migration was registered
2. Migration gets registered (tokens moved to IBC escrow)
3. User calls MsgFinalizeTokenWithdrawal
4. OPHost checks if token has migration info
5. If migrated: Transfer from IBC escrow to receiver
6. If not migrated: Fall back to bridge account withdrawal
7. User receives L1 tokens (same as before)
```

## Module Documentation

- **[OPChild Module](opchild_module.md)**: L2 token operations, L2→L1 migration, and IBC→L2 conversion
- **[OPHost Module](ophost_module.md)**: L1 bridge coordination, L1→L2 migration via IBC transfer, and in-flight withdrawal handling from IBC escrow
- **[IBC Middleware](ibc_middleware.md)**: Intercepts IBC packets and triggers IBC→L2 conversion
- **[Technical Specification](technical_specification.md)**: Detailed technical implementation and flows
- **[Flow Diagrams](flow_diagrams.md)**: Visual representations of the system

## Technical Implementation

### IBC Middleware Integration

The IBC middleware intercepts incoming IBC transfer packets and automatically triggers IBC→L2 conversion when:

1. The IBC denom is registered in the IBC→L2 denom map
2. The underlying `OnRecvPacket` call succeeds
3. The receiver's balance for the IBC denom increases after processing (i.e. additional IBC tokens arrived)

### OPHost IBC Transfer Integration

OPHost processes `MsgInitiateTokenDeposit` by:

1. Checking migration info for the bridge and L1 denom
2. Encoding bridge hook data in IBC transfer memo
3. Creating and routing IBC transfer messages
4. Preserving all existing functionality

### OPHost In-Flight Withdrawal Handling

OPHost processes `MsgFinalizeTokenWithdrawal` for in-flight requests by:

1. Checking migration info for the bridge and L1 denom
2. If migrated: Transfer tokens from IBC escrow to receiver
3. If not migrated: Fall back to bridge account withdrawal
4. Ensuring seamless handling of requests initiated before migration

### OPChild Token Conversion

OPChild handles:

1. **L2→IBC**: Explicit token migration via `MsgMigrateToken`
2. **IBC→L2**: Automatic conversion via `HandleMigratedTokenDeposit`
3. **L2→L1**: Withdrawal with automatic migration via `MsgInitiateTokenWithdrawal`
4. **Standard IBC**: Normal IBC transfers via `MsgTransfer` (no conversion)
