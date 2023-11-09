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
)
