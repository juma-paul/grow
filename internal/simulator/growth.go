package simulator

type GrowthStrategy interface {
	NextCapacity(needed int) int
}

// CPythonGrowth implements GrowthStrategy using CPython's exact formula.
// Source: Objects/listobject.c → list_resize()
type CPythonGrowth struct{}

func (CPythonGrowth) NextCapacity(needed int) int {
	unaligned := needed + (needed >> 3) + 6
	return unaligned &^ 3
}
