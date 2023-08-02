module minitia_std::op_bridge {
    use std::signer;
    use std::error;
    use std::string::{String};
    use std::event::{Self, EventHandle};

    use minitia_std::type_info;
    use minitia_std::coin::{Self, Coin};
    use minitia_std::table::{Self, Table};

    struct BridgeStore<phantom CoinType> has key {
        // The number is assigned for each bridge operations.
        sequence: u64,
        // Index all finalized deposit operations by the sequence number,
        // which is assigned from the l1 bridge.
        finalized_deposits: Table<u64, bool>,

        token_bridge_finalized_events: EventHandle<TokenBridgeFinalizedEvent>,
        token_bridge_initiated_events: EventHandle<TokenBridgeInitiatedEvent>,
    }

    //
    // Events
    //

    // Emitted when a token bridge is finalized on l2 chain.
    struct TokenBridgeFinalizedEvent has drop, store {
        from: address, // l1 address
        to: address, // l2 address
        l2_token: vector<u8>,
        amount: u64,
        l1_sequence: u64, // the sequence number which is assigned from the l1 bridge
    }

    // Emitted when a token bridge is initiated to the l1 chain.
    struct TokenBridgeInitiatedEvent has drop, store {
        from: address, // l2 address
        to: address, // l1 address
        l2_token: vector<u8>,
        amount: u64,
        l2_sequence: u64, // the operation sequence number
    }

    //
    // Errors
    //

    /// Only 0x1 is allowed to execute functions
    const EUNAUTHORIZED: u64 = 1;

    /// Duplicate register request
    const EALREADY_REGISTEREED: u64 = 2;

    /// The store is not registered
    const ENOT_REGISTERED: u64 = 3;

    /// The deposit tx is already finalized.
    const EALREADY_FINALIZED: u64 = 4;

    struct CapabilityStore<phantom CoinType> has key {
        burn_cap: coin::BurnCapability<CoinType>,
        freeze_cap: coin::FreezeCapability<CoinType>,
        mint_cap: coin::MintCapability<CoinType>,
    }

    fun assert_chain_permission(chain: &signer) {
        assert!(signer::address_of(chain) == @minitia_std, error::permission_denied(EUNAUTHORIZED));
    }

    /// Register a new token to bridge with coin initialization.
    public entry fun register_token<CoinType> (
        chain: &signer, 
        name: String,
        symbol: String,
        decimals: u8,
    ) {
        assert_chain_permission(chain);

        let chain_addr = signer::address_of(chain);
        assert!(!exists<BridgeStore<CoinType>>(chain_addr), error::already_exists(EALREADY_REGISTEREED));
        assert!(!exists<CapabilityStore<CoinType>>(chain_addr), error::already_exists(EALREADY_REGISTEREED));

        let (burn_cap, freeze_cap, mint_cap) = coin::initialize<CoinType>(
            chain,
            name,
            symbol,
            decimals,
        );

        move_to(chain, CapabilityStore<CoinType> {
            burn_cap,
            freeze_cap,
            mint_cap,
        });

        move_to(chain, BridgeStore<CoinType> {
            sequence: 0,
            finalized_deposits: table::new(),
            token_bridge_finalized_events: event::new_event_handle<TokenBridgeFinalizedEvent>(chain),
            token_bridge_initiated_events: event::new_event_handle<TokenBridgeInitiatedEvent>(chain),
        });
    }

    /// Finalize L1 => L2 token bridge operation.
    public entry fun finalize_token_bridge<CoinType> (
        chain: &signer,
        from: address,  // l1 sender address
        to: address,    // l2 receipient address
        amount: u64, 
        sequence: u64, // l1 bridge sequence number
    ) acquires CapabilityStore, BridgeStore {
        assert_chain_permission(chain);

        assert!(exists<BridgeStore<CoinType>>(@minitia_std), error::not_found(ENOT_REGISTERED));
        assert!(exists<CapabilityStore<CoinType>>(@minitia_std), error::not_found(ENOT_REGISTERED));

        let caps = borrow_global<CapabilityStore<CoinType>>(@minitia_std);
        let mint_coin = coin::mint<CoinType>(amount, &caps.mint_cap);
        coin::deposit<CoinType>(to, mint_coin);

        let l2_token = type_info::module_name(&type_info::type_of<CoinType>());
        let bridge_store = borrow_global_mut<BridgeStore<CoinType>>(@minitia_std);

        // check the deposit tx is already finalized.
        assert!(!table::contains(&bridge_store.finalized_deposits, sequence), error::invalid_state(EALREADY_FINALIZED));

        // index the sequence.
        table::add(&mut bridge_store.finalized_deposits, sequence, true);

        // emit event
        event::emit_event<TokenBridgeFinalizedEvent>(
            &mut bridge_store.token_bridge_finalized_events,
            TokenBridgeFinalizedEvent {
                from,
                to,
                l2_token,
                amount,
                l1_sequence: sequence,
            }
        )
    }

    /// User facing withdraw function to withdraw tokens from L2 to L1.
    public entry fun withdraw_token<CoinType> (
        account: &signer,
        to: address,
        amount: u64,
    ) acquires CapabilityStore, BridgeStore {
        initiate_token_bridge<CoinType>(signer::address_of(account), to, coin::withdraw(account, amount))
    }

    /// Initiate L2 => L1 withdraw bridge operation
    public fun initiate_token_bridge<CoinType> (
        from: address,
        to: address,
        amount: Coin<CoinType>,
    ) acquires CapabilityStore, BridgeStore {
        assert!(exists<BridgeStore<CoinType>>(@minitia_std), error::not_found(ENOT_REGISTERED));
        assert!(exists<CapabilityStore<CoinType>>(@minitia_std), error::not_found(ENOT_REGISTERED));

        // prepare event outputs
        let withdraw_amount = coin::value(&amount);
        let l2_token = type_info::module_name(&type_info::type_of<CoinType>());

        let caps = borrow_global<CapabilityStore<CoinType>>(@minitia_std);
        coin::burn<CoinType>(amount, &caps.burn_cap);

        // increase bridge operation sequence
        let bridge_store = borrow_global_mut<BridgeStore<CoinType>>(@minitia_std);
        bridge_store.sequence = bridge_store.sequence + 1;

        event::emit_event<TokenBridgeInitiatedEvent>(
            &mut bridge_store.token_bridge_initiated_events,
            TokenBridgeInitiatedEvent {
                from,
                to,
                l2_token,
                amount: withdraw_amount,
                l2_sequence: bridge_store.sequence,
            }
        )
    }

    #[test_only]
    use std::string;

    #[test_only]
    struct TestToken {}

    #[test(chain = @0x1)]
    fun test_resgier_coin(chain: &signer) {
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );
    }

    #[test(anonymous = @0x123)]
    #[expected_failure(abort_code = 0x50001, location = Self)]
    fun test_resgier_coin_permission(anonymous: &signer) {
        register_token<TestToken>(
            anonymous,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );
    }

    #[test(chain = @0x1)]
    #[expected_failure(abort_code = 0x80002, location = Self)]
    fun test_resgier_coin_multiple(chain: &signer) {
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );
    }

    #[test(chain = @0x1, from = @0x999, to = @0x998)]
    fun test_finalize_token_bridge(chain: &signer, from: address, to: address) acquires CapabilityStore, BridgeStore {
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );

        coin::init_module_for_test(chain);
        finalize_token_bridge<TestToken>(
            chain,
            from,
            to,
            1000000,
            1,
        );
    }

    #[test(chain = @0x1, from = @0x999, to = @0x998)]
    #[expected_failure(abort_code = 0x30004, location = Self)]
    fun test_finalize_token_bridge_failed_duplicate_sequence(chain: &signer, from: address, to: address) acquires CapabilityStore, BridgeStore {
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );

        coin::init_module_for_test(chain);
        finalize_token_bridge<TestToken>(
            chain,
            from,
            to,
            1000000,
            1,
        );

        finalize_token_bridge<TestToken>(
            chain,
            from,
            to,
            1000000,
            1,
        );
    }

    #[test(chain = @0x1, from = @0x999, to = @0x998)]
    fun test_initiate_token_bridge (chain: &signer, from: address, to: &signer) acquires CapabilityStore, BridgeStore {
        register_token<TestToken>(
            chain,
            string::utf8(b"test"),
            string::utf8(b"test"),
            8,
        );

        coin::init_module_for_test(chain);
        finalize_token_bridge<TestToken>(
            chain,
            from,
            signer::address_of(to),
            1000000,
            1,
        );

        coin::register<TestToken>(to);
        withdraw_token<TestToken>(
            to,
            from,
            1000000
        );

        let bridge_store = borrow_global<BridgeStore<TestToken>>(signer::address_of(chain));
        assert!(bridge_store.sequence == 1, 1);
    }
}