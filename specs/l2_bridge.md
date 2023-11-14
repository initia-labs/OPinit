# L2 Bridge

## Events

### `TokenBridgeFinalizedEvent`

The event is emitted when a executor finalized the token transfer from the L1 to L2.

```rust
// Emitted when a token bridge is finalized on l2 chain.
struct TokenBridgeFinalizedEvent has drop, store {
    from: address, // l1 address
    to: address, // l2 address
    l2_token: vector<u8>,
    amount: u64,
    l1_sequence: u64, // the sequence number which is assigned from the l1 bridge
}
```

### `TokenBridgeInitiatedEvent`

The event is emitted when a user executes `withdraw_token` function to move token from L2 to L1.

- The bridge module maintain `sequence` number to give unique identifier to each relay operation.
- In v1, `l2_sequence` is the unique identifier.

```rust
// Emitted when a token bridge is initiated to the l1 chain.
struct TokenBridgeInitiatedEvent has drop, store {
    from: address, // l2 address
    to: address, // l1 address
    l2_token: vector<u8>,
    amount: u64,
    l2_sequence: u64, // the operation sequence number
}
```

## Operations

### Register Token

This function allows the block executor to initialize a new token type with registration on the bridge module. The name of the newly deployed module should follow the L1 bridge contract’s event message `l2_token`, such as `01::l2_${l2_token}::Coin`.

### Finalize Token Bridge

This function finalizes the token transfer from L1 to L2. Only the block executor is allowed to execute this operation.

### Initiate Token Bridge

This function initiates the token bridge from L2 to L1. Users can execute `withdraw_token` to send tokens from L2 to L1. This operation emits the `TokenBridgeInitiatedEvent` with an `l2_sequence` number to prevent duplicate execution on L1.

The block executor should monitor this event to build withdraw storage for withdrawal proofs.
