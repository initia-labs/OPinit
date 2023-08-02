module initia_std::op_bridge {
    use std::signer;
    use std::string::{Self, String};
    use std::event::{Self, EventHandle};
    use std::hash::sha3_256;
    use std::vector;
    use std::error;

    use initia_std::coin::{Self, Coin};
    use initia_std::table::{Self, Table};
    use initia_std::type_info;
    use initia_std::bcs;

    use initia_std::op_output;

    //
    // Data Types
    //

    struct BridgeStore<phantom L2ID> has key {
        // The number is assigned for each bridge operations.
        sequence: u64,
        // L2 token => l1 token mapping.
        token_map: Table<vector<u8>, String>,

        token_registered_events: EventHandle<TokenRegisteredEvent>,
        token_bridge_initiated_events: EventHandle<TokenBridgeInitiatedEvent>,
        token_bridge_finalized_events: EventHandle<TokenBridgeFinalizedEvent>,
    }

    struct DepositStore<phantom L2ID, phantom CoinType> has key {
        deposits: Coin<CoinType>,
        // Index all proven withdrawals by its merkle value.
        proven_withdrawals: Table<vector<u8>, bool>,
    }

    //
    // Events
    //

    /// Emitted when deposit store is registered.
    struct TokenRegisteredEvent has drop, store {
        l2_id: String,
        l1_token: String,
        l2_token: vector<u8>, // sha3_256(type_name(`L2ID`) || type_name(`l1_token`))
    }

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

    //
    // Errors
    //

    /// l2 id already exists.
    const EBRIDGE_ALREADY_EXISTS: u64 = 1;

    /// l2 id not exists.
    const EBRIDGE_NOT_EXISTS: u64 = 2;

    /// Deposit store already exists.
    const EDEPOSIT_STORE_ALREADY_EXISTS: u64 = 3;

    /// Deposit store not exists.
    const EDEPOSIT_STORE_NOT_EXISTS: u64 = 4;

    /// Address of account which is used to initialize a bridge of `L2ID` doesn't match the deployer of module.
    const EL2_ADDRESS_MISMATCH: u64 = 5;

    /// Failed to generate same `output_root` with output root proofs.
    const EINVALID_OUTPUT_ROOT_PROOFS: u64 = 6;

    /// Proof length must be 32.
    const EINVALID_PROOF_LEGNTH: u64 = 7;

    /// Failed to generate same `storage_root` with the given withdrawal proofs.
    const EINVALID_STORAGE_ROOT_PROOFS: u64 = 8;

    /// The l2_ouput is not finalized yet.
    const EOUTPUT_NOT_FINALIZED: u64 = 9;

    /// The withdrawal tx is already proved and claimed.
    const EALREADY_PROVED: u64 = 10;

    //
    // Heldper Functions
    //

    /// A helper function that returns the address of L2ID.
    fun l2_address<L2ID>(): address {
        let type_info = type_info::type_of<L2ID>();
        type_info::account_address(&type_info)
    }

    #[view]
    /// A helper function that returns l2 token name bytes
    public fun l2_token<L2ID, CoinType>(): vector<u8> {
        let l2_id = type_info::type_name<L2ID>();
        let l1_token = type_info::type_name<CoinType>();
        
        // l2_token = sha3_256(l2_id || l1_token)
        let l2_token_seed = *string::bytes(&l2_id);
        vector::append<u8>(&mut l2_token_seed, *string::bytes(&l1_token));
        sha3_256(l2_token_seed)
    }

    /// 0: equal
    /// 1: v1 is greator than v2
    /// 2: v1 is less than v2
    fun bytes_cmp(v1: &vector<u8>, v2: &vector<u8>): u8 {
        assert!(vector::length(v1) == 32, error::invalid_argument(EINVALID_PROOF_LEGNTH));
        assert!(vector::length(v2) == 32, error::invalid_argument(EINVALID_PROOF_LEGNTH));

        let i = 0;
        while (i < 32 ) {
            let e1 = *vector::borrow(v1, i);
            let e2 = *vector::borrow(v2, i);
            if (e1 > e2) {
                return 1
            } else if (e2 > e1) {
                return 2
            };
        };

        0
    }

    //
    // Entry Functions
    //

    /// create bridge store
    public entry fun initialize<L2ID>(account: &signer) {
        let account_addr = signer::address_of(account);
        assert!(!exists<BridgeStore<L2ID>>(account_addr), error::already_exists(EBRIDGE_ALREADY_EXISTS));
        assert!(l2_address<L2ID>() == account_addr, error::permission_denied(EL2_ADDRESS_MISMATCH));

        // register new bridge store
        move_to(account, BridgeStore<L2ID> {
            sequence: 0,
            token_map: table::new(),
            token_registered_events: event::new_event_handle<TokenRegisteredEvent>(account),
            token_bridge_initiated_events: event::new_event_handle<TokenBridgeInitiatedEvent>(account),
            token_bridge_finalized_events: event::new_event_handle<TokenBridgeFinalizedEvent>(account),
        });
    }

    /// register coin to bridge store and prepare deposit store
    public entry fun register_token<L2ID, CoinType>(account: &signer) acquires BridgeStore {
        let account_addr = signer::address_of(account);
        assert!(l2_address<L2ID>() == account_addr, error::permission_denied(EL2_ADDRESS_MISMATCH));
        assert!(!exists<DepositStore<L2ID, CoinType>>(account_addr), error::already_exists(EDEPOSIT_STORE_ALREADY_EXISTS));

        // register new deposit store
        move_to(account, DepositStore<L2ID, CoinType> {
            deposits: coin::zero(),
            proven_withdrawals: table::new(),
        });

        // prepare event outputs
        let l2_id = type_info::type_name<L2ID>();
        let l1_token = type_info::type_name<CoinType>();
        let l2_token = l2_token<L2ID, CoinType>();
        
        // add token mapping
        let bridge_store = borrow_global_mut<BridgeStore<L2ID>>(l2_address<L2ID>());
        table::add(&mut bridge_store.token_map, l2_token, l1_token);

        // emit event
        event::emit_event<TokenRegisteredEvent>(
            &mut bridge_store.token_registered_events,
            TokenRegisteredEvent {
                l2_id,
                l1_token,
                l2_token,
            } 
        );
    }

    /// user facing l2 deposit function
    public entry fun deposit_token<L2ID, CoinType>(account: &signer, to: address, amount: u64) acquires BridgeStore, DepositStore {
        initiate_token_bridge<L2ID, CoinType>(signer::address_of(account), to, coin::withdraw<CoinType>(account, amount))
    }

    /// initiate l1 => l2 deposit bridge operation
    public fun initiate_token_bridge<L2ID, CoinType>(from: address, to: address, amount: Coin<CoinType>) acquires BridgeStore, DepositStore {
        let l2_addr = l2_address<L2ID>();
        assert!(exists<BridgeStore<L2ID>>(l2_addr), error::not_found(EBRIDGE_NOT_EXISTS));
        assert!(exists<DepositStore<L2ID, CoinType>>(l2_addr), error::not_found(EDEPOSIT_STORE_NOT_EXISTS));

        // prepare event outputs
        let l2_id = type_info::type_name<L2ID>();
        let l1_token = type_info::type_name<CoinType>();
        let l2_token = l2_token<L2ID, CoinType>();
        let deposit_amount = coin::value<CoinType>(&amount);

        let bridge_store = borrow_global_mut<BridgeStore<L2ID>>(l2_addr);
        let deposit_store = borrow_global_mut<DepositStore<L2ID, CoinType>>(l2_addr);
        coin::merge<CoinType>(&mut deposit_store.deposits, amount);
        bridge_store.sequence = bridge_store.sequence + 1;

        // emit event
        event::emit_event<TokenBridgeInitiatedEvent>(
            &mut bridge_store.token_bridge_initiated_events,
            TokenBridgeInitiatedEvent {
                from,
                to,
                l2_id,
                l1_token,
                l2_token,
                amount: deposit_amount,
                l1_sequence: bridge_store.sequence,
            }
        );
    }

    // prove withdraw transation and withdraw the token to 
    // the receiver address
    public entry fun finalize_token_bridge<L2ID, CoinType>(
        l2_output_index: u64,
        withdrawal_proofs: vector<vector<u8>>,
        // withdraw tx data
        sequence: u64,          // sequence which is assigned from l2's bridge contract 
        sender: address,        // address of the sender of the transaction
        receiver: address,      // address of the receiver of the transaction
        amount: u64,            // amount to send to the reciepient
        // output root proofs
        version: vector<u8>,            // version of the output root
        state_root: vector<u8>,         // l2 state root
        storage_root: vector<u8>,       // withdrawal state root
        lastest_block_hash: vector<u8>, // l2 latest block hash
    ) acquires BridgeStore, DepositStore {
        assert!(op_output::is_finalized<L2ID>(l2_output_index), error::invalid_state(EOUTPUT_NOT_FINALIZED));

        // validate output root generation
        {
            let output_root_seed = vector::empty<u8>();
            vector::append(&mut output_root_seed, version);
            vector::append(&mut output_root_seed, state_root);
            vector::append(&mut output_root_seed, storage_root);
            vector::append(&mut output_root_seed, lastest_block_hash);
            let output_root = sha3_256(output_root_seed);
            
            // check output root proof validation
            assert!(output_root == op_output::get_output_root<L2ID>(l2_output_index), error::invalid_argument(EINVALID_OUTPUT_ROOT_PROOFS));
        };
        
        let l2_addr = l2_address<L2ID>();
        let deposit_store = borrow_global_mut<DepositStore<L2ID, CoinType>>(l2_addr);

        // verify storage root can be generated with
        // withdrawal proofs and withdraw tx data
        {
            // convert withdraw tx data into hash
            let withdrawal_hash = {
                let withdraw_tx_data = vector::empty<u8>();
                vector::append(&mut withdraw_tx_data, bcs::to_bytes(&sequence));
                vector::append(&mut withdraw_tx_data, bcs::to_bytes(&sender));
                vector::append(&mut withdraw_tx_data, bcs::to_bytes(&receiver));
                vector::append(&mut withdraw_tx_data, bcs::to_bytes(&amount));
                vector::append(&mut withdraw_tx_data, *string::bytes(&type_info::type_name<CoinType>()));
                
                sha3_256(withdraw_tx_data)
            };

            // check already proved 
            assert!(!table::contains(&deposit_store.proven_withdrawals, withdrawal_hash), EALREADY_PROVED);

            // should works with sorted merkle tree
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

            // add the withdrawal_hash to proven list
            table::add(&mut deposit_store.proven_withdrawals, withdrawal_hash, true);
        };

        
        let withdrawn_coin = coin::extract<CoinType>(&mut deposit_store.deposits, amount);

        // deposit the withdrawn coin to receiver address
        coin::deposit<CoinType>(receiver, withdrawn_coin);

        // prepare event outputs
        let from = sender;
        let to = receiver;
        let l2_id = type_info::type_name<L2ID>();
        let l1_token = type_info::type_name<CoinType>();
        let l2_token = l2_token<L2ID, CoinType>();

        let bridge_store = borrow_global_mut<BridgeStore<L2ID>>(l2_addr);
        event::emit_event<TokenBridgeFinalizedEvent>(
            &mut bridge_store.token_bridge_finalized_events,
            TokenBridgeFinalizedEvent {
                from,
                to,
                l2_id, 
                l1_token,
                l2_token,
                amount,
                l2_sequence: sequence,
            }
        )
    }

    #[test_only]
    struct GameID{}

    #[test_only]
    use std::coin::{BurnCapability, FreezeCapability, MintCapability};

    #[test_only]
    use std::block;

    #[test_only]
    use initia_std::native_uinit::Coin as Token;

    #[test_only]
    struct TestCapabilityStore<phantom CoinType> has key {
        burn_cap: BurnCapability<CoinType>,
        freeze_cap: FreezeCapability<CoinType>,
        mint_cap: MintCapability<CoinType>,
    }

    #[test_only]
    public fun test_setup(
        chain: &signer, 
        proposer: address, 
        challenger: address, 
        mint_amount: u64,
    ): Coin<Token> acquires BridgeStore {

        // initialize coin
        coin::init_module_for_test(chain);
        let (burn_cap, freeze_cap, mint_cap) = coin::initialize<Token>(
            chain,
            string::utf8(b"INIT Coin"),
            string::utf8(b"uinit"),
            6,
        );

        let mint_coin = coin::mint<Token>(mint_amount, &mint_cap);

        move_to(chain, TestCapabilityStore<Token> {
            burn_cap,
            freeze_cap,
            mint_cap,
        });

        // initialize bridge
        initialize<GameID>(chain);
        register_token<GameID, Token>(chain);

        // initialize output
        op_output::initialize<GameID>(
            chain,
            100,
            proposer,
            challenger,
            100,
            100,
        );

        block::set_block_info(100, 100);

        mint_coin
    }

    #[test(chain=@0x1, proposer=@0x998, challenger=@0x997, from=@0x996, to=@0x995)]
    fun verify_merkle_proof(
        chain: &signer, 
        proposer: &signer, 
        challenger: &signer, 
        from: &signer, 
        to: &signer,
    ) acquires BridgeStore, DepositStore{
        let test_coin = test_setup(chain, signer::address_of(proposer), signer::address_of(challenger), 10000000000);

        initiate_token_bridge<GameID, Token>(signer::address_of(from), signer::address_of(to), test_coin);

        let output_root_seed = vector::empty<u8>();
        vector::append(&mut output_root_seed, x"1234123412341234123412341234123412341234123412341234123412341234");
        vector::append(&mut output_root_seed, x"4321432143214321432143214321432143214321432143214321432143214321");
        vector::append(&mut output_root_seed, x"40a33f52cae58feeaacfc20332bfa07f3476b77dcec6c96b7686bb3a4ce67e61");
        vector::append(&mut output_root_seed, x"9999999999999999999999999999999999999999999999999999999999999999");
        let output_root = sha3_256(output_root_seed);
        assert!(output_root == x"4b3d6ff89837ba5b61f6a372c29b96ba0bdad9eb013efa9431d4ef5e2b1a8fc2", 1);

        op_output::propose_l2_output<GameID>(proposer, output_root, 100);

        // update block info to finalize
        block::set_block_info(200, 200);

        let withdrawal_proofs: vector<vector<u8>> = vector::empty();
        vector::push_back(&mut withdrawal_proofs, x"984098c883030cfaddca39d195edf037e16fdb7d4c78a61c0302382274e7b6c8");
        vector::push_back(&mut withdrawal_proofs, x"c56fa652ddd91da50f2a74c7de122943f04bd4c398af0eb61758550bcb77e8bb");
        vector::push_back(&mut withdrawal_proofs, x"5c927e5bdffca99f01a2afbbe80b18f082d771cf879a33fe67806afb87ba7afb");

        coin::register<Token>(to);
        finalize_token_bridge<GameID, Token>(
            0,
            withdrawal_proofs,
            101,
            signer::address_of(from),
            signer::address_of(to),
            1000001,
            x"1234123412341234123412341234123412341234123412341234123412341234",
            x"4321432143214321432143214321432143214321432143214321432143214321",
            x"40a33f52cae58feeaacfc20332bfa07f3476b77dcec6c96b7686bb3a4ce67e61",
            x"9999999999999999999999999999999999999999999999999999999999999999",
        );

        assert!(coin::balance<Token>(signer::address_of(to)) == 1000001, 2);
    }
}