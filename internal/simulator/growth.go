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

// DoublingGrowth implements the classic textbook doubling strategy.
// Capacity is always a power of 2, minimum 4.
type DoublingGrowth struct{}

func (DoublingGrowth) NextCapacity(needed int) int {
	capacity := 4
	for capacity < needed {
		capacity *= 2
	}
	return capacity
}

// OneAndAHalfGrowth implements Java ArrayList-style 1.5× growth.
// Result is ceil(needed * 1.5), minimum 4.
type OneAndAHalfGrowth struct{}

func (OneAndAHalfGrowth) NextCapacity(needed int) int {
	grown := needed + (needed+1)/2
	return max(4, grown)
}
