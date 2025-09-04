# Bridge Replacement System Flow Diagrams

## Overview

This document provides visual representations of the bridge replacement system flows, showing how the system transitions from OP Bridge to IBC Bridge while preserving user experience.

## System Architecture Diagram

```mermaid
graph TB
    subgraph "L1 Chain (OPHost)"
        A[User calls MsgInitiateTokenDeposit]
        B[OPHost Module]
        C[IBC Transfer Creation]
    end
    
    subgraph "IBC Network"
        D[IBC Transfer]
    end
    
    subgraph "L2 Chain (OPChild)"
        E[IBC Middleware]
        F[OPChild Module]
        G[User receives OP tokens]
    end
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    
    style A fill:#e1f5fe
    style G fill:#e1f5fe
    style B fill:#f3e5f5
    style F fill:#f3e5f5
    style E fill:#fff3e0
```

## Bridge Replacement Flow

### L1 → L2 Flow (Two Options)

#### Option A: Bridge Replacement (User Experience Preserved)

```mermaid
sequenceDiagram
    participant User
    participant OPHost
    participant IBC
    participant Middleware
    participant OPChild
    
    User->>OPHost: MsgInitiateTokenDeposit
    OPHost->>OPHost: Check migration info
    OPHost->>IBC: Create IBC transfer
    IBC->>Middleware: IBC packet received
    Middleware->>OPChild: HandleMigratedTokenDeposit
    OPChild->>OPChild: Burn IBC tokens
    OPChild->>OPChild: Mint OP tokens
    OPChild->>User: Send OP tokens
    Note over User: User gets OP tokens, not IBC tokens
```

#### Option B: Normal IBC Bridge

```mermaid
sequenceDiagram
    participant User
    participant IBC
    participant L2
    
    User->>IBC: MsgTransfer (INIT)
    IBC->>L2: IBC transfer (IBC INIT)
    L2->>L2: IBC middleware converts IBC INIT to OP INIT
    L2->>User: OP INIT tokens
    Note over User: User gets OP INIT tokens (converted from IBC INIT)
```

### L2 → L1 Flow (Two Options)

#### Option A: Automatic Migration

```mermaid
sequenceDiagram
    participant User
    participant OPChild
    participant IBC
    participant L1
    
    User->>OPChild: MsgInitiateTokenWithdrawal
    OPChild->>OPChild: Convert OP to IBC
    OPChild->>IBC: IBC transfer
    IBC->>L1: IBC tokens received
    Note over L1: IBC tokens = L1 tokens (e.g., INIT)
```

#### Option B: Explicit Migration

```mermaid
sequenceDiagram
    participant User
    participant OPChild
    participant IBC
    participant L1
    
    User->>OPChild: MsgMigrateToken
    OPChild->>User: IBC tokens
    User->>IBC: Manual IBC transfer
    IBC->>L1: IBC tokens received
    Note over L1: IBC tokens = L1 tokens (e.g., INIT)
```

## IBC Middleware Integration

### Packet Interception Flow

```mermaid
flowchart TD
    A[IBC Packet Received] --> B{Parse Transfer Data}
    B -->|Success| C{Check Source Chain}
    B -->|Fail| D[Pass to Standard IBC]
    C -->|Not Source| E[Compute IBC Denom]
    C -->|Source| D
    E --> F{Check Migration Info}
    F -->|Not Found| D
    F -->|Found| G[Record Pre-Balance]
    G --> H[Process Standard IBC]
    H --> I{Check Balance Change}
    I -->|No Change| J[Return Success]
    I -->|Increased| K[Trigger Conversion]
    K --> L[HandleMigratedTokenDeposit]
    L --> M[Emit Events]
    M --> J
```

## Bridge Hook Preservation Flow

```mermaid
sequenceDiagram
    participant User
    participant OPHost
    participant IBC
    participant Middleware
    participant OPChild
    
    User->>OPHost: MsgInitiateTokenDeposit + Hook Data
    OPHost->>OPHost: Encode hook in memo
    OPHost->>IBC: IBC transfer with memo
    IBC->>Middleware: IBC packet with hook data
    Middleware->>OPChild: HandleMigratedTokenDeposit
    OPChild->>OPChild: Decode hook data
    OPChild->>OPChild: Execute hook
    OPChild->>User: OP tokens + Hook results
```

## In-Flight Withdrawal Handling on OPHost (Initiated Before Migration)

```mermaid
sequenceDiagram
    participant User
    participant OPHost
    participant IBC Escrow
    
    Note over User, IBC Escrow: In-Flight Withdrawal Request Handling
    User->>OPHost: MsgFinalizeTokenWithdrawal (initiated before migration)
    OPHost->>OPHost: Check migration info
    alt Token is migrated
        OPHost->>OPHost: Get IBC escrow address
        OPHost->>IBC Escrow: Transfer tokens
        IBC Escrow->>User: L1 tokens
        Note over User: Withdrawal from IBC escrow
    else Token is not migrated
        OPHost->>OPHost: Bridge account withdrawal
        OPHost->>User: L1 tokens
        Note over User: Normal bridge withdrawal
    end
```
