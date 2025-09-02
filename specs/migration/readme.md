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

### Bridge Replacement Flows

#### L1 → L2 Flow (Two Options)

**Option A: Bridge Replacement (User Experience Preserved)**

```plaintext
User calls MsgInitiateTokenDeposit → Receives OP tokens on L2
```

**Option B: Normal IBC Bridge**

```plaintext
User calls MsgTransfer → IBC transfer → L2 receives IBC tokens
```

#### L2 → L1 Flow (Two Options)

**Option A: Automatic Migration (User Experience Preserved)**

```plaintext
User calls MsgInitiateTokenWithdrawal → Receives L1 tokens
```

**Option B: Explicit Migration**

```plaintext
User calls MsgMigrateToken → Gets IBC tokens → Manual IBC transfer
```

#### Bridge Hook Preservation

```plaintext
1. User calls MsgInitiateTokenDeposit with hook data
2. OPHost encodes hook data as MigratedTokenDepositMemo
3. Hook data travels via IBC transfer memo
4. OPChild receives IBC tokens and decodes hook data
5. OPChild executes hook using existing hook function
6. User gets OP tokens + hook execution results
```

## Module Documentation

- **[OPChild Module](opchild_module.md)**: L2 token operations, L2→L1 migration, and IBC→L2 conversion
- **[OPHost Module](ophost_module.md)**: L1 bridge coordination and L1→L2 migration via IBC transfer
- **[IBC Middleware](ibc_middleware.md)**: Intercepts IBC packets and triggers IBC→L2 conversion
- **[Technical Specification](technical_specification.md)**: Detailed technical implementation and flows
- **[Flow Diagrams](flow_diagrams.md)**: Visual representations of the system

## Technical Implementation

### IBC Middleware Integration

The IBC middleware intercepts incoming IBC transfer packets and automatically triggers IBC→L2 conversion when:

1. IBC tokens are received for registered denoms
2. Balance increases after IBC processing
3. Migration info exists for the IBC denom

### OPHost IBC Transfer Integration

OPHost processes `MsgInitiateTokenDeposit` by:

1. Checking migration info for the bridge and L1 denom
2. Encoding bridge hook data in IBC transfer memo
3. Creating and routing IBC transfer messages
4. Preserving all existing functionality

### OPChild Token Conversion

OPChild handles:

1. **L2→IBC**: Explicit token migration via `MsgMigrateToken`
2. **IBC→L2**: Automatic conversion via `HandleMigratedTokenDeposit`
3. **L2→L1**: Withdrawal with automatic migration via `MsgInitiateTokenWithdrawal`
4. **Standard IBC**: Normal IBC transfers via `MsgTransfer` (no conversion)
