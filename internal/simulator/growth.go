package simulator

// Interface (contract)
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

// Doubling implementation
type DoublingGrowth struct{}

func (DoublingGrowth) NextCapacity(needed int) int {
	capacity := 4
	for capacity < needed {
		capacity *= 2
	}
	return capacity
}