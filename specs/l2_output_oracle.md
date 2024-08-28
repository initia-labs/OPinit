# L2 Output Oracle

In version 1, output oracle maintain `proposer` and `challenger` addresses on its config store. The `proposer` can submit the `output_proposal` and the `challenger` can delete the output when the proposed output state is wrong.

The first version of the implementation does not include a dispute system, but uses permissioned propose and challenge mechanisms. In version 2, anyone can propose the output with a certain amount of `stake`, and disputes will be resolved based on the governance of L1.

## Operations

### Propose L2 Output

L2 output oracle receives `output_root` with L2 block number to check the checkpoint of L2. The checkpoints are the multiple of `submission_interval`. A proposer must submit the `output_root` at the every checkpoints.

The followings are the components of `output_root`.

- `version`: the version of output root
- `state_root`: l2 state root
- `storage_root`: withdrawal storage root
- `last_block_hash`: l2 latest block hash

To build the `output_root`, concatenate all the components in sequence and apply `sha3_256`.

### Delete L2 Output

A challenger can delete the output without dispute in version 1 with output index.

### Update Proposer

The operation is to update proposer to another address when a proposer keeps submitting a invalid output root. The operation is supposed to be executed by `0x1` via L1 governance.

### Update Challenger

The operation is to update challenger to another address when a challenger keeps deleting a valid output root. The operation is supposed to be executed by `0x1` via L1 governance.
