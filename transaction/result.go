package transaction

// Result represents a command result index.
type Result struct {
	Index uint16
}

// Arg returns the result as an Argument.
func (r Result) Arg() Argument {
	idx := r.Index
	return Argument{Result: &idx}
}

// At returns the nested result at the provided index.
func (r Result) At(resultIndex uint16) Argument {
	idx := r.Index
	res := resultIndex
	return Argument{NestedResult: &NestedResult{Index: idx, ResultIndex: res}}
}

// Arguments returns nested results for the provided count.
func (r Result) Arguments(count int) []Argument {
	if count <= 0 {
		return nil
	}

	args := make([]Argument, count)
	for i := 0; i < count; i++ {
		args[i] = r.At(uint16(i))
	}

	return args
}
