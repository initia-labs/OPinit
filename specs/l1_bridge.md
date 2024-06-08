# L1 Bridge Module

## Events

### `TokenRegisteredEvent`

The event is emitted when a new token support is added to the bridge contract.

* In v1 spec, the bridge executor should add a new token support manually.

```rust
/// Emitted when deposit store is registered.
struct TokenRegisteredEvent has drop, store {
    l2_id: String,
    l1_token: String,
    l2_token: vector<u8>, // sha3_256(type_name(`L2ID`) || type_name(`l1_token`))
}
```

### `TokenBridgeInitiatedEvent`

The event is emitted when a user executes `deposit_token` function to move a token from L1 to L2.

* The bridge module maintains `sequence` number to give a unique identifier for each relaying operation.
* In v1, `l2_id` + `l1_sequence` is the unique identifier.

```rust
/// Emitted when a token bridge is initiated to the l2 chain.
struct TokenBridgeInitiatedEvent has drop, store {
    from: address, // l1 address
    to: address, // l2 address
    l2_id: String,
    l1_token: String,
    l2_token: vector<u8>,
    amount: u64,
    l1_sequence: u64, 
}
```

### `TokenBridgeFinalizedEvent`

The event is emitted when a withdrawal transaction is proved and finalized.
  
* In v1, `sha3(bcs(l2_sequence) + bcs(from) + bcs(to) + bcs(amount) + bytes(l1_token))` is the unique identifier for each withdrawal operation.

```rust
/// Emitted when a token bridge is finalized on l1 chain.
struct TokenBridgeFinalizedEvent has drop, store {
    from: address, // l2 address
    to: address, // l1 address
    l2_id: String,
    l1_token: String,
    l2_token: vector<u8>,
    amount: u64,
    l2_sequence: u64, // the sequence number which is assigned from the l2 bridge
}
```

## Operations

### Register Token

This function is for the bridge executor, who controls the bridge relaying operations. In version 1, only the bridge executor can add support for a new token type. After registration, the `TokenRegisteredEvent` event is emitted to deploy a new coin type module on L2.

The bridge executor should monitor the `TokenRegisteredEvent` and deploy a new coin type module on L2. They should also execute the L2 bridge's `/minitia.rollup.v1.MsgCreateToken` function for initialization.

### Initiate Token Bridge

This function enables a user to transfer their asset from L1 to L2. The deposited token will be locked in the bridge's `DepositStore` and can only be released using the `withdraw` operation in L2. When executed, this operation emits a `TokenBridgeInitiatedEvent`. A bridge executor should subscribe to this event and transfer the token to L2 by executing the `finalize_token_bridge` function in the L2 bridge.

### Finalize Token Bridge

This function is used to prove and finalize the withdrawal transaction from L2. The proving process is described [here](https://www.notion.so/Withdrawal-Proving-a49f7c26467044489731048f68ed584b?pvs=21). Once the proving is complete, the deposited tokens are withdrawn to the recipient address, and the `TokenBridgeFinalizedEvent` event is emitted. To prevent duplicate withdrawal attempts, the bridge uses a unique identifier calculated as `sha3(bcs(l2_sequence) + bcs(from) + bcs(to) + bcs(amount) + bytes(l1_token))`.
