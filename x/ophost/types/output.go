package types

func (output Output) Validate() error {
	if len(output.OutputRoot) != 32 {
		return ErrInvalidHashLength.Wrap("output_root")
	}

	return nil
}

func (output Output) IsEmpty() bool {
	return len(output.OutputRoot) == 0 && output.L1BlockTime.IsZero() && output.L2BlockNumber == 0
}
