package types

type ExecutorChangePlan struct {
	// L1 governance proposal id
	ProposalID    int64
	Height        int64
	NextExecutor  string
	NextValidator Validator
	Info          string
}
