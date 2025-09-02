# IBC Middleware Migration System

## Overview

The IBC Middleware Migration system provides seamless integration between the OPChild migration module and the Inter-Blockchain Communication (IBC) protocol. It automatically intercepts incoming IBC transfer packets and triggers token conversion when appropriate, enabling **bridge replacement** from OP Bridge to IBC Bridge while ensuring users receive OP INIT tokens instead of IBC INIT tokens, maintaining the same user experience.

## Architecture

### Core Components

#### `IBCMiddleware` Struct

- **Purpose**: Wraps underlying IBC modules and provides automatic IBC transfer handling
- **Interfaces**: Implements `porttypes.Middleware` and `porttypes.UpgradableModule`
- **Dependencies**: Bank keeper, OPChild keeper, address codec, and underlying IBC app

#### Key Fields

- **`ac`**: Address codec for address conversion
- **`app`**: Underlying IBC module for standard operations
- **`ics4Wrapper`**: ICS4 wrapper for packet capabilities
- **`bankKeeper`**: Bank module keeper for balance operations
- **`opChildKeeper`**: OPChild module keeper for migration logic

## Core Functionality

### 1. Packet Interception

#### `OnRecvPacket`

- **Purpose**: Intercepts incoming IBC transfer packets
- **Process**:
  1. **Packet Validation**: Validates transfer packet data
  2. **Source Chain Check**: Identifies if token originated from receiving chain
  3. **IBC Denom Computation**: Calculates IBC denom from packet data
  4. **Migration Check**: Verifies if token is registered for conversion
  5. **Balance Tracking**: Records pre-transfer balance
  6. **Standard Processing**: Calls underlying IBC module
  7. **Token Conversion**: Executes IBC to L2 conversion if balance increased
  8. **Event Emission**: Emits conversion events

#### Packet Processing Flow

```plain
Incoming IBC Packet → Validate Transfer Data → Check Source Chain → 
Compute IBC Denom → Check Conversion Registration → Record Pre-Balance → 
Process Standard IBC → Check Balance Change → Trigger Token Conversion → 
Emit Events → Return Acknowledgement
```

### 2. Automatic Token Conversion

#### IBC Transfer Handling

- **Trigger**: Balance increase after IBC transfer
- **Process**: Automatic call to `HandleMigratedTokenDeposit`
- **Result**: IBC tokens burned, L2 tokens minted
- **User Experience**: Seamless automatic conversion of received IBC tokens

#### Conversion Conditions

- **Token Registration**: Must be registered in `IBCToL2DenomMap`
- **Balance Increase**: Post-transfer balance must exceed pre-transfer
- **Valid Packet**: IBC packet must be successfully processed

## Migration Integration

### 1. OPChild Module Integration

#### Keeper Interface

- **`HasIBCToL2DenomMap`**: IBC transfer conversion registration check
- **`HandleMigratedTokenDeposit`**: Core IBC transfer handling execution
- **Error Handling**: Comprehensive error management

#### IBC Transfer Flow

```plaintext
IBC Token Received → Check Conversion Registration → 
Execute Conversion → Burn IBC Token → Mint L2 Token → 
Handle Bridge Hooks → Emit Events → Return Success
```

## Bridge Hook Support

### 1. Hook Execution

#### Memo Parsing

- **Format**: JSON-encoded bridge hook data
- **Validation**: Strict parsing with error handling
- **Fallback**: Graceful handling of invalid memos

#### Hook Processing

- **OPinit Integration**: Execute OPinit bridge hooks
- **Error Handling**: Graceful hook failure handling

### 2. Hook Features

- **Custom Logic**: User-defined hook execution
- **Gas Control**: Configurable execution limits
- **State Integration**: Seamless state management
