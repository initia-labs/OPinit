package types

const (
	EventTypeRecordBatch             = "record_batch"
	EventTypeCreateBridge            = "create_bridge"
	EventTypeProposeOutput           = "propose_output"
	EventTypeDeleteOutput            = "delete_output"
	EventTypeInitiateTokenDeposit    = "initiate_token_deposit"
	EventTypeFinalizeTokenWithdrawal = "finalize_token_withdrawal"
	EventTypeUpdateProposer          = "update_proposer"
	EventTypeUpdateChallenger        = "update_challenger"
	EventTypeOraclePrice             = "oracle_price"

	AttributeKeySubmitter     = "submitter"
	AttributeKeyCreator       = "creator"
	AttributeKeyProposer      = "proposer"
	AttributeKeyChallenger    = "challenger"
	AttributeKeyBridgeId      = "bridge_id"
	AttributeKeyOutputIndex   = "output_index"
	AttributeKeyOutputRoot    = "output_root"
	AttributeKeyL2BlockNumber = "l2_block_number"
	AttributeKeyFrom          = "from"
	AttributeKeyTo            = "to"
	AttributeKeyAmount        = "amount"
	AttributeKeyL1Denom       = "l1_denom"
	AttributeKeyL2Denom       = "l2_denom"
	AttributeKeyData          = "data"
	AttributeKeyL1Sequence    = "l1_sequence"
	AttributeKeyL2Sequence    = "l2_sequence"

	// oracle
	AttributeKeyBase        = "base"
	AttributeKeyQuote       = "quote"
	AttributeKeyPrice       = "price"
	AttributeKeyBlockTime   = "block_time"
	AttributeKeyBlockHeight = "block_height"
)
