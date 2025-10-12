package migration

const (
	EventTypeHandleMigratedTokenDeposit = "handle_migrated_token_deposit"
	EventTypeHandleMigratedTokenRefund  = "handle_migrated_token_refund"
	AttributeKeyReceiver                = "receiver"
	AttributeKeyAmount                  = "amount"
	AttributeKeyIbcDenom                = "ibc_denom"
)
