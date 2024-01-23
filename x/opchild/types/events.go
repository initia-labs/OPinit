package types

// Rollup module event types
const (
	EventTypeAddValidator            = "add_validator"
	EventTypeRemoveValidator         = "remove_validator"
	EventTypeFinalizeTokenDeposit    = "finalize_token_deposit"
	EventTypeInitiateTokenWithdrawal = "initiate_token_withdrawal"
	EventTypeExecuteMessages         = "execute_messages"
	EventTypeWhitelist               = "whitelist"
	EventTypeParams                  = "params"
	EventTypeOraclePrice             = "oracle_price"

	AttributeKeySender         = "sender"
	AttributeKeyRecipient      = "recipient"
	AttributeKeyAmount         = "amount"
	AttributeKeyDenom          = "denom"
	AttributeKeyStructTag      = "struct_tag"
	AttributeKeyValidator      = "validator"
	AttributeKeyMetadata       = "metadata"
	AttributeKeyL1Sequence     = "l1_sequence"
	AttributeKeyFinalizeHeight = "finalize_height"
	AttributeKeyHookSuccess    = "hook_success"
	AttributeKeyFrom           = "from"
	AttributeKeyTo             = "to"
	AttributeKeyL2Sequence     = "l2_sequence"

	// oracle
	AttributeKeyBase          = "base"
	AttributeKeyQuote         = "quote"
	AttributeKeyPrice         = "price"
	AttributeKeyL1BlockTime   = "l1_block_time"
	AttributeKeyL1BlockHeight = "l1_block_height"
)
