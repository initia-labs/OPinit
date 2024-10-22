# Launch Tools

Launch Tools is used to simplify initial setup of minitia.

## How to use

Prepare BridgeExecutor account which is funded L1 gas tokens.

```shell
export TARGET_NETWORK=mahalo-3

# Using default setup
#
# To use default setup, you need to prepare L1 coin funded
# bridge-executor account.
#
# Get funds from the faucet
# https://faucet.$TARGET_NETWORK.initia.xyz/
#
minitiad launch $TARGET_NETWORK

# using custom setup
minitiad launch $TARGET_NETWORK --with-config [path-to-config]
```

## Example of Config

```json
{
  "l1_config": {
    "chain_id": "initiation-2",
    "rpc_url": "https://rpc.initiation-2.initia.xyz:443",
    "gas_prices": "0.015uinit"
  },
  "l2_config": {
    "chain_id": "minitia-XDzNjd-1",
    "denom": "umin",
    "moniker": "operator"
  },
  "op_bridge": {
    "output_submission_start_height": 1,
    "output_submission_interval": "1h0m0s",
    "output_finalization_period": "168h0m0s",
    "batch_submission_target": "INITIA",
    "enable_oracle": true
  },
  "system_keys": {
    "validator": {
      "l2_address": "init12z54lfqgp7zapzuuk2m4h6mjz84qzca8j0wm4x",
      "mnemonic": "digital kingdom slim fall cereal aspect expose trade once antique treat spatial unfair trip silver diesel other friend invest valve human blouse decrease salt"
    },
    "bridge_executor": {
      "l1_address": "init13skjgs2x96c4sk9mfkfdzjywm75l6wy63j5gyn",
      "l2_address": "init13skjgs2x96c4sk9mfkfdzjywm75l6wy63j5gyn",
      "mnemonic": "junk aunt group member rebel dinosaur will trial jacket core club obscure morning unit fame round render napkin boy chest same patrol twelve medal"
    },
    "output_submitter": {
      "l1_address": "init1f4lu0ze9c7zegrrjfpymjvztucqz48z3cy8p5f"
    },
    "batch_submitter": {
      "da_address": "init1hqv5xqt7lckdj9p5kfp2q5auc5z37p2vyt4d72"
    },
    "challenger": {
      "l1_address": "init1gn0yjtcma92y27c0z84ratxf6juy69lpln6u88",
      "l2_address": "init1gn0yjtcma92y27c0z84ratxf6juy69lpln6u88"
    }
  },
  "genesis_accounts": [
    {
      "address": "init12z54lfqgp7zapzuuk2m4h6mjz84qzca8j0wm4x",
      "coins": "100000000umin"
    },
    {
      "address": "init13skjgs2x96c4sk9mfkfdzjywm75l6wy63j5gyn"
    },
    {
      "address": "init1gn0yjtcma92y27c0z84ratxf6juy69lpln6u88"
    }
  ]
}
```
