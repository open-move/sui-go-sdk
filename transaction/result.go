package transaction

type Result struct {
	Index uint16
}

func (r Result) Arg() Argument {
	idx := r.Index
	return Argument{Result: &idx}
}

func (r Result) At(resultIndex uint16) Argument {
	idx := r.Index
	res := resultIndex
	return Argument{NestedResult: &NestedResult{Index: idx, ResultIndex: res}}
}

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
