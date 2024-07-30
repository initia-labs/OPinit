# Bridge Hooks

## Bridge Hook

The Bridge Hook feature is similar to IBC hooks and allows for the execution of registered virtual machines (VMs). It enables cross-chain contract calls that involve token movement, making it useful for various use cases.

The mechanism behind Bridge Hooks is the `data` field on every `MsgInitiateTokenDeposit` message. When the `to` field is set to `hook` and the `data` field follows a specific format, a registered hook function is executed, providing execution guarantees.

### Intermediate Sender

To prevent unexpected error cases or attacks, we utilize the intermediate sender method. This method involves transferring the bridged asset to the intermediate sender and executing the bridge hook with the signer of the intermediate sender. This protection ensures that the bridge hook cannot spend more than the transferred amount.

### Execution Format

The bridge hook data format varies depending on the chain, allowing each chain to define its own format.

#### Move Chains

Move chains typically receive `MsgExecute` or `MsgExecuteJSON` messages.

To distinguish between these messages, you can include an `is_json` field in the data JSON message.

For `MsgExecute`, the JSON message would look like this:

```json
{
    ...,
    "is_json": false
}
```

And for `MsgExecuteJSON`, the JSON message would look like this:

```json
{
    ...,
    "is_json": true
}
```

This way, you can easily identify the type of message being sent.

```golang
type MsgExecute struct {
    Sender        string   `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
    ModuleAddress string   `protobuf:"bytes,2,opt,name=module_address,json=moduleAddress,proto3" json:"module_address,omitempty"`
    ModuleName    string   `protobuf:"bytes,3,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
    FunctionName  string   `protobuf:"bytes,4,opt,name=function_name,json=functionName,proto3" json:"function_name,omitempty"`
    TypeArgs      []string `protobuf:"bytes,5,rep,name=type_args,json=typeArgs,proto3" json:"type_args,omitempty"`
    Args          [][]byte `protobuf:"bytes,6,rep,name=args,proto3" json:"args,omitempty"`
}
```

or

```golang
// MsgExecuteJSON is the message to execute the given module function
type MsgExecuteJSON struct {
  // Sender is the that actor that signed the messages
  Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
  // ModuleAddr is the address of the module deployer
  ModuleAddress string `protobuf:"bytes,2,opt,name=module_address,json=moduleAddress,proto3" json:"module_address,omitempty"`
  // ModuleName is the name of module to execute
  ModuleName string `protobuf:"bytes,3,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
  // FunctionName is the name of a function to execute
  FunctionName string `protobuf:"bytes,4,opt,name=function_name,json=functionName,proto3" json:"function_name,omitempty"`
  // TypeArgs is the type arguments of a function to execute
  // ex) "0x1::BasicCoin::Initia", "bool", "u8", "u64"
  TypeArgs []string `protobuf:"bytes,5,rep,name=type_args,json=typeArgs,proto3" json:"type_args,omitempty"`
  // Args is the arguments of a function to execute in json stringify format
  Args []string `protobuf:"bytes,6,rep,name=args,proto3" json:"args,omitempty"`
}
```

#### Wasm Chains

Wasm chains typically receive `MsgExecuteContract` messages.

```golang
type MsgExecuteContract struct {
    Sender   string                                  `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
    Contract string                                  `protobuf:"bytes,2,opt,name=contract,proto3" json:"contract,omitempty"`
    Msg      RawContractMessage                      `protobuf:"bytes,3,opt,name=msg,proto3,casttype=RawContractMessage" json:"msg,omitempty"`
    Funds    github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,5,rep,name=funds,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"funds"`
}
```

#### EVM Chains

EVM chains typically receive `MsgCall` messages.

```golang
type MsgCall struct {
    Sender       string                  `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
    ContractAddr string                  `protobuf:"bytes,2,opt,name=contract_addr,json=contractAddr,proto3" json:"contract_addr,omitempty"`
    Input        string                  `protobuf:"bytes,3,opt,name=input,proto3" json:"input,omitempty"`
    Value        cosmossdk_io_math.Int   `protobuf:"bytes,4,opt,name=value,proto3,customtype=cosmossdk.io/math.Int" json:"value"`
}
```
