package types

func (output Output) Validate() error {
	if len(output.OutputRoot) != 32 {
		return ErrInvalidHashLength.Wrap("output_root")
	}

	return nil
}
