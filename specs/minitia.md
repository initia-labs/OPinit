# Minitia

## Messages

There are three categories of message in `x/rollup`` module.

* Bridge Executor messages
  * [`MsgCreateToken`](#msgcreatetoken)
  * [`MsgDeposit`](#msgdeposit)
* Validator messages
  * [`MsgExecuteMessages`](#msgexecutemessages)
  * [`MsgExecuteLegacyContents`](#msgexecutelegacycontents)
* Authority messages
  * [`MsgAddValidator`](#msgaddvalidator)
  * [`MsgRemoveValidator`](#msgremovevalidator)
  * [`MsgUpdateParams`](#msgupdateparams)
  * [`MsgWhitelist`](#msgwhitelist)
  * [`MsgSpendFeePool`](#msgspendfeepool)

### `MsgCreateToken`

The message is for a bridge executor to publish a new coin struct tag `0x1::native_${denom}::Coin` and initialize a new coin with that struct tag.

```proto
// MsgCreateToken is the message to create a new token from L1
message MsgCreateToken {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "rollup/MsgCreateToken";

  // the sender address
  string sender = 1 [
    (gogoproto.moretags) = "yaml:\"sender\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // denom is the denom of the token to create.
  string denom   = 2;
  string name    = 3;
  string symbol  = 4;
  int64 decimals = 5;
}
```

### `MsgDeposit`

The message is for a bridge executor to finalize a deposit request from L1. The message handler internally executes [`finalize_token_bridge`](./l2_bridge.md#finalize-token-bridge) of `l2_bridge`.

```proto
// MsgDeposit is the message to submit deposit funds from upper layer
message MsgDeposit {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "rollup/MsgDeposit";

  // the sender address
  string sender = 1 [
    (gogoproto.moretags) = "yaml:\"sender\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // from is l1 sender address
  string from = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // to is l2 recipient address
  string to = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // amount is the coin amount to deposit.
  cosmos.base.v1beta1.Coin amount = 4 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false,
    (amino.dont_omitempty) = true
  ];

  // sequence is the sequence number of l1 bridge
  uint64 sequence = 5;
}
```

### `MsgExecuteMessages`

The message is to execute authority messages with validator permission like `x/gov` module of cosmos-sdk. Any validator can execute the message with various authority messages.

```proto
// MsgExecuteMessages is the message to execute the given
// authority messages with validator permission.
message MsgExecuteMessages {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "rollup/MsgExecuteMessages";

  // Sender is the that actor that signed the messages
  string sender = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // messages are the arbitrary messages to be executed.
  repeated google.protobuf.Any messages = 2;
}
```

### `MsgExecuteLegacyContents`

The message is also copied from `x/gov` module of cosmos-sdk to support legacy param update of ibc modules. The execution permission is given to validators.

```proto

// MsgExecuteLegacyContents is the message to execute legacy
// (gov) contents with validator permission.
message MsgExecuteLegacyContents {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "rollup/MsgExecuteLegacyContents";

  // Sender is the that actor that signed the messages
  string sender = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // contents are the arbitrary legacy (gov) contents to be executed.
  repeated google.protobuf.Any contents = 2;
}
```

### `MsgAddValidator`

The message is to add a new validator to the comet-bft validator set. The execution permission is given to authority, which is `rollup` module account.

```proto
// MsgAddValidator defines a SDK message for adding a new validator.
message MsgAddValidator {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "rollup/MsgAddValidator";

  option (gogoproto.equal) = false;
  option (gogoproto.goproto_getters) = false;

  // authority is the address that controls the module
  // (defaults to x/rollup unless overwritten).
  string authority = 1 [
    (gogoproto.moretags) = "yaml:\"authority\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  string moniker = 2;
  string validator_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  google.protobuf.Any pubkey = 4 [(cosmos_proto.accepts_interface) = "cosmos.crypto.PubKey"];
}
```

### `MsgRemoveValidator`

The message is to remove a validator from the comet-bft validator set. The execution permission is given to authority, which is `rollup` module account.

```proto
// MsgAddValidator is the message to remove a validator from designated list
message MsgRemoveValidator {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "rollup/MsgRemoveValidator";

  // authority is the address that controls the module
  // (defaults to x/rollup unless overwritten).
  string authority = 1 [
    (gogoproto.moretags) = "yaml:\"authority\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // validator is the validator to remove.
  string validator_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### `MsgUpdateParams`

The message is to update the rollup module params. The execution permission is given to authority, which is `rollup` module account.

```proto
// MsgUpdateParams is the message to update legacy parameters
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "rollup/MsgUpdateParams";

  // authority is the address that controls the module
  // (defaults to x/rollup unless overwritten).
  string authority = 1 [
    (gogoproto.moretags) = "yaml:\"authority\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // params are the arbitrary parameters to be updated.
  Params params = 2;
}
```

### `MsgWhitelist`

The message is to add a coin type to whitelist for auto register. The execution permission is given to authority, which is `rollup` module account.

```proto
// whitelist a coin type to enable auto coin module register.
message MsgWhitelist {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "rollup/MsgWhitelist";

  // authority is the address that controls the module
  // (defaults to x/rollup unless overwritten).
  string authority = 1 [
    (gogoproto.moretags) = "yaml:\"authority\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // coin_type is the struct tag to whitelist.
  string coin_type = 2;
}
```

### `MsgSpendFeePool`

The message is to spend collected fee. The execution permission is given to authority, which is `rollup` module account.

```proto
// MsgSpendFeePool is the message to withdraw collected fees from the module account to the recipient address.
message MsgSpendFeePool {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "rollup/MsgSpendFeePool";

  // authority is the address that controls the module
  // (defaults to x/rollup unless overwritten).
  string authority = 1 [
    (gogoproto.moretags) = "yaml:\"authority\"",
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];

  // recipient is address to receive the coins.
  string recipient = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // the coin amount to spend.
  repeated cosmos.base.v1beta1.Coin amount = 3 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false,
    (amino.dont_omitempty) = true,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"
  ];
}
```
