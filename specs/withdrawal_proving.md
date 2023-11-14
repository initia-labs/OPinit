# Withdrawal Proving

`ominitia` defines `submission_interval`, which is the L2 block number at which checkpoints must be submitted. At each `submission_interval`, the bridge executor should build the withdraw storage, which is the Merkle Tree for the withdrawal verification process on L1.

`ominitia` uses a sorted Merkle Tree to reduce verifying cost, and each tree node is referred to as a `withdrawal_hash`.

The following are the components of `withdrawal_hash`:

- `sequence`: The L2 bridge sequence assigned to each withdrawal operation.
- `sender`: The address from which the withdrawal operation is made.
- `receiver`: The address to which the withdrawal operation is made.
- `amount`: The token amount of the withdrawal operation.
- `coin_type`: The L1 token struct tag.

To build the `withdrawal_hash`, concatenate all the components and apply `sha3_256` after serializing the values with `bcs`, except for `coin_type`, because the `coin_type` is already a string that can be converted to bytes.

```rust
fun verify(
    withdrawal_proofs: vector<vector<u8>>, 
    sequence: u64,
    sender: address,
    receiver: address,
    amount: u64,
    coin_type: String,
): bool {
    let withdrawal_hash = {
        let withdraw_tx_data = vector::empty<u8>();
      vector::append(&mut withdraw_tx_data, bcs::to_bytes(&sequence));
      vector::append(&mut withdraw_tx_data, bcs::to_bytes(&sender));
      vector::append(&mut withdraw_tx_data, bcs::to_bytes(&receiver));
      vector::append(&mut withdraw_tx_data, bcs::to_bytes(&amount));
      vector::append(&mut withdraw_tx_data, *string::bytes(&type_info::type_name<CoinType>()));
      
      sha3_256(withdraw_tx_data)
    };

    let i = 0;
    let len = vector::length(&withdrawal_proofs);
    let root_seed = withdrawal_hash;
    while (i < len) {
        let proof = vector::borrow(&withdrawal_proofs, i);
        let cmp = bytes_cmp(&root_seed, proof);
        root_seed = if (cmp == 2 /* less */) {
            let tmp = vector::empty();
            vector::append(&mut tmp, root_seed);
            vector::append(&mut tmp, *proof);

            sha3_256(tmp)
        } else /* greator or equals */ {
            let tmp = vector::empty();
            vector::append(&mut tmp, *proof);
            vector::append(&mut tmp, root_seed);

            sha3_256(tmp)
        };
        i = i + 1;
    };

    let root_hash = root_seed;
    assert!(storage_root == root_hash, error::invalid_argument(EINVALID_STORAGE_ROOT_PROOFS));
}
```

The example implementation of building the Merkle Tree can be found [here](https://github.com/initia-labs/op-bridge-executor).
