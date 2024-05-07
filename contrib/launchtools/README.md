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
 "l1-config": {
  "chain-id": "mahalo-3",
  "rpc-url": "https://rpc.mahalo-3.initia.xyz:443",
  "gas-prices": "0.15uinit"
 },
 "l2-config": {
  "chain-id": "minitia-PBkigP-1",
  "denom": "umin",
  "moniker": "operator"
 },
 "op-bridge": {
  "output-submission-start-time": "2024-05-07T18:31:24.575116+09:00",
  "output-submission-interval": 3600000000000,
  "output-finalization-period": 3600000000000,
  "batch-submission-target": "l1"
 },
 "system-keys": {
  "validator": {
   "address": "init1sms9vlcf0cn04lt2vuw8246eesgsr3unqcrx7p",
   "mnemonic": "update cradle typical salute chunk utility confirm ghost science mad bonus ring word cube copy negative endless exact fan artwork pulp chunk valid melody"
  },
  "bridge-executor": {
   "address": "init13skjgs2x96c4sk9mfkfdzjywm75l6wy63j5gyn",
   "mnemonic": "junk aunt group member rebel dinosaur will trial jacket core club obscure morning unit fame round render napkin boy chest same patrol twelve medal"
  },
  "output-submitter": {
   "address": "init1xz3769tse6kj2nr6d0u3cucyjkayee3arhc4f5",
   "mnemonic": "rug wealth pull physical dragon zebra crash poem weird also exist click return benefit garbage omit mandate letter hidden decline uncle glare found cupboard"
  },
  "batch-submitter": {
   "address": "init19um6segpjw38gjuuuh57pwpl0dutzw70h8k88k",
   "mnemonic": "very motor food appear stone purse eyebrow round current rare install humble giant assume lawn giraffe tray lecture aspect people weird razor wrestle viable"
  },
  "challenger": {
   "address": "init17kyjcs03d2avxn039f96jcfracahdk2sw40p82",
   "mnemonic": "surprise foil sibling leisure ocean sketch north cover crowd tiger extra hair laugh hint small sight pole mistake unlock borrow river corn shell measure"
  }
 },
 "genesis-accounts": [
  {
   "address": "init1sms9vlcf0cn04lt2vuw8246eesgsr3unqcrx7p",
   "coins": "1000000000umin"
  },
  {
   "address": "init13skjgs2x96c4sk9mfkfdzjywm75l6wy63j5gyn"
  },
  {
   "address": "init1xz3769tse6kj2nr6d0u3cucyjkayee3arhc4f5"
  },
  {
   "address": "init19um6segpjw38gjuuuh57pwpl0dutzw70h8k88k"
  },
  {
   "address": "init17kyjcs03d2avxn039f96jcfracahdk2sw40p82"
  }
 ]
}
```
