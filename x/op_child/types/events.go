package types

// Rollup module event types
const (
	EventTypeAddValidator        = "add_validator"
	EventTypeRemoveValidator     = "remove_validator"
	EventTypeFinalizeTokenBridge = "finalize_token_bridge"
	EventTypeInitiateTokenBridge = "initiate_token_bridge"
	EventTypeExecuteMessages     = "execute_messages"
	EventTypeWhitelist           = "whitelist"
	EventTypeParams              = "params"

	AttributeKeySender           = "sender"
	AttributeKeyRecipient        = "recipient"
	AttributeKeyAmount           = "amount"
	AttributeKeyDenom            = "denom"
	AttributeKeyStructTag        = "struct_tag"
	AttributeKeyValidator        = "validator"
	AttributeKeyMetadata         = "metadata"
	AttributeKeyInboundSequence  = "inbound_sequence"
	AttributeKeyFinalizeHeight   = "finalize_height"
	AttributeKeyHookSuccess      = "hook_success"
	AttributeKeyFrom             = "from"
	AttributeKeyTo               = "to"
	AttributeKeyOutboundSequence = "outbound_sequence"
)
