package types

type ExecutorChangePlan struct {
	// L1 governance proposal id
	ProposalID uint64

	// Upgrade height
	Height uint64

	// Next executor addresses
	NextExecutors []string

	// Next validator
	NextValidator Validator

	// Additional information
	Info string
}
