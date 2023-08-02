module initia_std::op_output {
    use std::event::{Self, EventHandle};
    use std::signer;
    use std::option;
    use std::vector;
    use std::error;
    
    use initia_std::table::{Self, Table};
    use initia_std::type_info::{Self};
    use initia_std::string::{String};
    use initia_std::block;

    //
    // Data Types
    //

    struct ConfigStore<phantom L2ID> has copy, drop, key {
        // The interval in L2 blocks at which checkpoints must be submitted.
        submission_interval: u64, 
        // The address of the challenger.
        challenger: address,
        // The address of the proposer.
        proposer: address,
        // The minimum time (in seconds) that must elapse before a withdrawal can be finalized.
        finalization_period_seconds: u64,
        // The number of the first L2 block recorded in this contract.
        starting_block_number: u64,
    }

    struct OutputStore<phantom L2ID> has key {
        outputs: Table<u64, OutputProposal<L2ID>>,
        output_proposed_events: EventHandle<OutputProposedEvent>,
        output_deleted_events: EventHandle<OutputDeletedEvent>,
    }

    struct OutputProposal<phantom L2ID> has copy, store, drop {
        // Hash of the L2 output.
        output_root: vector<u8>,
        // Timestamp of the L1 block that the output root was submitted in.
        l1_timestamp: u64,
        // L2 block number that the output corresponds to.
        l2_block_number: u64,
    }

    //
    // Events
    //

    // Emitted when output is proposed.
    struct OutputProposedEvent has store, drop {
        // The id of L2.
        l2_id: String,
        // The output root.
        output_root: vector<u8>,
        // The index of the output in the l2_outputs array.
        output_index: u64,
        // The L2 block number of the output root.address
        l2_block_number: u64,
        // The L1 timestamp when proposed.
        l1_timestamp: u64,
    }

    // Emitted when outputs are deleted.
    struct OutputDeletedEvent has store, drop {
        // The id of L2.
        l2_id: String,
        // Next L2 output index before the deletion.
        prev_next_output_index: u64,
        // Next L2 output index after the deletion.
        new_next_output_index: u64,
    }

    //
    // Errors
    //

    /// Config store already exists.
    const ECONFIG_STORE_ALREADY_EXISTS: u64 = 1;

    /// Config store not exists.
    const ECONFIG_STORE_NOT_EXISTS: u64 = 2;

    /// Output store already exists.
    const EOUTPUT_STORE_ALREADY_EXISTS: u64 = 3;

    /// Output store not exists.
    const EOUTPUT_STORE_NOT_EXISTS: u64 = 4;

    /// Address of account which is used to initialize a output of `L2ID` doesn't match the deployer of module.
    const EL2_ADDRESS_MISMATCH: u64 = 5;

    /// Proposer address is not matched with the registered one in the config.
    const EPROPOSER_ADDRESS_MISMATCH: u64 = 6;

    /// Challenger address is not matched with the registered one in the config.
    const ECHALLENGER_ADDRESS_MISMATCH: u64 = 7;

    /// Block number must be equal to next expected block number.
    const EL2_BLOCK_NUM_MISMATCH: u64 = 8;

    /// Hash bytes vector length should be 32.
    const EINVALID_HASH_LENGTH: u64 = 9;

    /// Out of output index.
    const EOUT_OF_OUTPUT_INDEX: u64 = 10;

    /// Fialized output cannot be deleted.
    const EOUTPUT_FINALIZED: u64 = 11;

    /// Chain permission assertion.
    const EASSERT_CHAIN_PERMISSION: u64 = 12;

    /// Unknown errors.
    const EUNKNOWN: u64 = 99;

    //
    // Heldper Functions
    //

    /// A helper function that returns the address of L2ID.
    fun l2_address<L2ID>(): address {
        let type_info = type_info::type_of<L2ID>();
        type_info::account_address(&type_info)
    }

    //
    // View Functions
    //

    #[view]
    public fun get_config_store<L2ID>(): ConfigStore<L2ID> acquires ConfigStore {
        let l2_addr = l2_address<L2ID>();
        let config_store = borrow_global<ConfigStore<L2ID>>(l2_addr);
        *config_store
    }

    #[view]
    public fun get_output_root<L2ID>(output_index: u64): vector<u8> acquires OutputStore {
        let output_proposal = get_output_proposal<L2ID>(output_index);
        output_proposal.output_root
    }

    #[view]
    public fun is_finalized<L2ID>(output_index: u64): bool acquires ConfigStore, OutputStore {
        let l2_addr = l2_address<L2ID>();
        let config_store = borrow_global<ConfigStore<L2ID>>(l2_addr);
        let (_, block_timestamp) = block::get_block_info();

        let output_proposal = get_output_proposal<L2ID>(output_index);
        block_timestamp >= config_store.finalization_period_seconds + output_proposal.l1_timestamp
    }

    #[view]
    public fun get_output_proposal<L2ID>(output_index: u64): OutputProposal<L2ID> acquires OutputStore {
        let l2_addr = l2_address<L2ID>();
        let output_store = borrow_global<OutputStore<L2ID>>(l2_addr);

        let output_proposal = table::borrow(&output_store.outputs, output_index);
        *output_proposal
    }

    #[view]
    public fun next_block_num<L2ID>(): u64 acquires ConfigStore, OutputStore {
        let l2_addr = l2_address<L2ID>();
        let config_store = borrow_global<ConfigStore<L2ID>>(l2_addr);
        let output_store = borrow_global<OutputStore<L2ID>>(l2_addr);

        let next_output_index = table::length(&output_store.outputs);
        let next_block_num = if (next_output_index == 0) {
            config_store.starting_block_number
        } else {
            let iter = table::iter(&output_store.outputs, option::none(), option::none(), 2);
            assert!(table::prepare<u64, OutputProposal<L2ID>>(&mut iter), error::aborted(EUNKNOWN));
            let (_, output_proposal) = table::next<u64, OutputProposal<L2ID>>(&mut iter);

            config_store.submission_interval + output_proposal.l2_block_number
        };

        next_block_num
    }

    #[view]
    public fun next_output_index<L2ID>(): u64 acquires OutputStore {
        let l2_addr = l2_address<L2ID>();
        let output_store = borrow_global<OutputStore<L2ID>>(l2_addr);
        table::length(&output_store.outputs)
    }

    //
    // Entry Functions
    //

    /// Update challenger to another address.
    /// Permission is granted to 0x1 to delegate decision-making 
    /// authorization for challenge disputes to L1 governance.
    public entry fun update_challenger<L2ID> (
        challenger: &signer,
        new_challenger: address,
    ) acquires ConfigStore {
        let l2_addr = l2_address<L2ID>();
        assert!(exists<ConfigStore<L2ID>>(l2_addr), error::not_found(ECONFIG_STORE_NOT_EXISTS));

        let config_store = borrow_global_mut<ConfigStore<L2ID>>(l2_addr);
        assert!(signer::address_of(challenger) == @initia_std, error::unauthenticated(EASSERT_CHAIN_PERMISSION));

        config_store.challenger = new_challenger;
    }

    /// Update proposer to another address.
    /// Permission is granted to 0x1 to delegate decision-making 
    /// authorization for challenge disputes to L1 governance.
    public entry fun update_proposer<L2ID> (
        proposer: &signer,
        new_proposer: address,
    ) acquires ConfigStore {
        let l2_addr = l2_address<L2ID>();
        assert!(exists<ConfigStore<L2ID>>(l2_addr), error::not_found(ECONFIG_STORE_NOT_EXISTS));

        let config_store = borrow_global_mut<ConfigStore<L2ID>>(l2_addr);
        assert!(signer::address_of(proposer) == @initia_std, error::unauthenticated(EASSERT_CHAIN_PERMISSION));

        config_store.proposer = new_proposer;
    }

    /// create output store
    public entry fun initialize<L2ID> (
        account: &signer,
        submission_interval: u64,
        proposer: address,
        challenger: address,
        finalization_period_seconds: u64,
        starting_block_number: u64,
    ) {
        let account_addr = signer::address_of(account);
        assert!(!exists<ConfigStore<L2ID>>(account_addr), error::already_exists(ECONFIG_STORE_ALREADY_EXISTS));
        assert!(!exists<OutputStore<L2ID>>(account_addr), error::already_exists(EOUTPUT_STORE_ALREADY_EXISTS));
        assert!(l2_address<L2ID>() == account_addr, error::invalid_argument(EL2_ADDRESS_MISMATCH));

        // register new bridge store
        move_to(account, ConfigStore<L2ID> {
            submission_interval,
            challenger,
            proposer,
            finalization_period_seconds,
            starting_block_number,
        });

        move_to(account, OutputStore<L2ID>{
            outputs: table::new(),
            output_proposed_events: event::new_event_handle<OutputProposedEvent>(account),
            output_deleted_events: event::new_event_handle<OutputDeletedEvent>(account),
        });
    }

    /// TODO - allow anyone to propose with stake
    public entry fun propose_l2_output<L2ID>(
        account: &signer,
        output_root: vector<u8>,
        l2_block_number: u64,
    ) acquires ConfigStore, OutputStore {        
        let l2_addr = l2_address<L2ID>();
        let config_store = borrow_global<ConfigStore<L2ID>>(l2_addr);
        assert!(signer::address_of(account) == config_store.proposer, error::unauthenticated(EPROPOSER_ADDRESS_MISMATCH));

        let output_store = borrow_global_mut<OutputStore<L2ID>>(l2_addr);
        let next_output_index = table::length(&output_store.outputs);
        let next_block_num = if (next_output_index == 0) {
            config_store.starting_block_number
        } else {
            let iter = table::iter(&output_store.outputs, option::none(), option::none(), 2);
            assert!(table::prepare<u64, OutputProposal<L2ID>>(&mut iter), error::aborted(EUNKNOWN));
            let (_, output_proposal) = table::next<u64, OutputProposal<L2ID>>(&mut iter);

            config_store.submission_interval + output_proposal.l2_block_number
        };

        assert!(l2_block_number == next_block_num, error::invalid_argument(EL2_BLOCK_NUM_MISMATCH));
        assert!(vector::length(&output_root) == 32, error::invalid_argument(EINVALID_HASH_LENGTH));

        // store output proposal
        let (_, l1_timestamp) = block::get_block_info();
        table::add<u64, OutputProposal<L2ID>>(
            &mut output_store.outputs, 
            next_output_index, 
            OutputProposal {
                output_root,
                l1_timestamp,
                l2_block_number,
            }
        );

        // emit proposed event
        let l2_id = type_info::type_name<L2ID>();
        event::emit_event<OutputProposedEvent>(
            &mut output_store.output_proposed_events,
            OutputProposedEvent {
                l2_id,
                output_root,
                output_index: next_output_index,
                l2_block_number,
                l1_timestamp,
            } 
        );
    }

    /// Delete L2 output proposal. Only challenger is allowed to execute
    /// the function.
    public entry fun delete_l2_output<L2ID>(
        account: &signer,
        output_index: u64,
    ) acquires ConfigStore, OutputStore {
        assert!(!is_finalized<L2ID>(output_index), error::invalid_argument(EOUTPUT_FINALIZED));

        let l2_addr = l2_address<L2ID>();
        let config_store = borrow_global<ConfigStore<L2ID>>(l2_addr);
        assert!(signer::address_of(account) == config_store.challenger, error::unauthenticated(ECHALLENGER_ADDRESS_MISMATCH));

        let output_store = borrow_global_mut<OutputStore<L2ID>>(l2_addr);
        let next_output_index = table::length(&output_store.outputs);
        assert!(output_index < next_output_index, error::invalid_argument(EOUT_OF_OUTPUT_INDEX));
        while (output_index < next_output_index) {
            table::remove<u64, OutputProposal<L2ID>>(&mut output_store.outputs, output_index);
            output_index = output_index+1;
        };

        let l2_id = type_info::type_name<L2ID>();
        event::emit_event<OutputDeletedEvent>(
            &mut output_store.output_deleted_events,
            OutputDeletedEvent {
                l2_id,
                prev_next_output_index: next_output_index,
                new_next_output_index: output_index,
            }
        );
    }

    #[test_only]
    struct TestL2ID {}

    #[test_only]
    use std::block::set_block_info;

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    fun test_initialize(chain: &signer, proposer: address, challenger: address) acquires ConfigStore {
        initialize<TestL2ID>(chain, 100, proposer, challenger, 200, 300);

        let config = get_config_store<TestL2ID>();
        assert!(config.submission_interval == 100, 0);
        assert!(config.challenger == challenger, 0);
        assert!(config.proposer == proposer, 0);
        assert!(config.finalization_period_seconds == 200, 0);
        assert!(config.starting_block_number == 300, 0);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3, new_challenger=@0x4)]
    fun test_update_challenger(chain: &signer, proposer: address, challenger: &signer, new_challenger: address) acquires ConfigStore {
        initialize<TestL2ID>(chain, 100, proposer, signer::address_of(challenger), 200, 300);

        let config = get_config_store<TestL2ID>();
        assert!(config.submission_interval == 100, 0);
        assert!(config.challenger == signer::address_of(challenger), 0);
        assert!(config.proposer == proposer, 0);
        assert!(config.finalization_period_seconds == 200, 0);
        assert!(config.starting_block_number == 300, 0);

        update_challenger<TestL2ID>(chain, new_challenger);
        let config = get_config_store<TestL2ID>();
        assert!(config.challenger == new_challenger, 0);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3, new_challenger=@0x4)]
    #[expected_failure(abort_code = 0x4000C, location = Self)]
    fun test_failed_update_challenger(chain: &signer, proposer: address, challenger: &signer, new_challenger: address) acquires ConfigStore {
        initialize<TestL2ID>(chain, 100, proposer, signer::address_of(challenger), 200, 300);

        let config = get_config_store<TestL2ID>();
        assert!(config.submission_interval == 100, 0);
        assert!(config.challenger == signer::address_of(challenger), 0);
        assert!(config.proposer == proposer, 0);
        assert!(config.finalization_period_seconds == 200, 0);
        assert!(config.starting_block_number == 300, 0);

        update_challenger<TestL2ID>(challenger, new_challenger);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3, new_proposer=@0x4)]
    fun test_update_proposer(chain: &signer, proposer: &signer, challenger: address, new_proposer: address) acquires ConfigStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), challenger, 200, 300);

        let config = get_config_store<TestL2ID>();
        assert!(config.submission_interval == 100, 0);
        assert!(config.challenger == challenger, 0);
        assert!(config.proposer == signer::address_of(proposer), 0);
        assert!(config.finalization_period_seconds == 200, 0);
        assert!(config.starting_block_number == 300, 0);

        update_proposer<TestL2ID>(chain, new_proposer);
        let config = get_config_store<TestL2ID>();
        assert!(config.proposer == new_proposer, 0);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3, new_proposer=@0x4)]
    #[expected_failure(abort_code = 0x4000C, location = Self)]
    fun test_fail_update_proposer(chain: &signer, proposer: &signer, challenger: address, new_proposer: address) acquires ConfigStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), challenger, 200, 300);

        let config = get_config_store<TestL2ID>();
        assert!(config.submission_interval == 100, 0);
        assert!(config.challenger == challenger, 0);
        assert!(config.proposer == signer::address_of(proposer), 0);
        assert!(config.finalization_period_seconds == 200, 0);
        assert!(config.starting_block_number == 300, 0);

        update_proposer<TestL2ID>(proposer, new_proposer);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    fun test_propose_l2_output(chain: &signer, proposer: &signer, challenger: address) acquires ConfigStore, OutputStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), challenger, 200, 301);

        set_block_info(100, 123);
        propose_l2_output<TestL2ID>(proposer, x"0000000000000000000000000000000000000000000000000000000000000001", 301);

        let output_proposal = get_output_proposal<TestL2ID>(0);
        assert!(output_proposal.output_root ==  x"0000000000000000000000000000000000000000000000000000000000000001", 0);
        assert!(output_proposal.l2_block_number == 301, 0);
        assert!(output_proposal.l1_timestamp == 123, 0);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    #[expected_failure(abort_code = 0x40006, location = Self)]
    fun test_fail_unauthorized_propose_l2_output(chain: &signer, proposer: &signer, challenger: address) acquires ConfigStore, OutputStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), challenger, 200, 301);

        set_block_info(100, 123);
        propose_l2_output<TestL2ID>(chain, x"0000000000000000000000000000000000000000000000000000000000000001", 301);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    #[expected_failure(abort_code = 0x10008, location = Self)]
    fun test_fail_wrong_block_num_propose_l2_output(chain: &signer, proposer: &signer, challenger: address) acquires ConfigStore, OutputStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), challenger, 200, 301);

        set_block_info(100, 123);
        propose_l2_output<TestL2ID>(proposer, x"0000000000000000000000000000000000000000000000000000000000000001", 201);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    fun test_delete_l2_output(chain: &signer, proposer: &signer, challenger: &signer) acquires ConfigStore, OutputStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), signer::address_of(challenger), 200, 301);

        set_block_info(100, 123);
        propose_l2_output<TestL2ID>(proposer, x"0000000000000000000000000000000000000000000000000000000000000001", 301);

        delete_l2_output<TestL2ID>(challenger, 0);
    }

    #[test(chain=@0x1, proposer=@0x2, challenger=@0x3)]
    #[expected_failure(abort_code = 0x40007, location = Self)]
    fun test_fail_unauthorized_delete_l2_output(chain: &signer, proposer: &signer, challenger: &signer) acquires ConfigStore, OutputStore {
        initialize<TestL2ID>(chain, 100, signer::address_of(proposer), signer::address_of(challenger), 200, 301);

        set_block_info(100, 123);
        propose_l2_output<TestL2ID>(proposer, x"0000000000000000000000000000000000000000000000000000000000000001", 301);

        delete_l2_output<TestL2ID>(chain, 0);
    }
}