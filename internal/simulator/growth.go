package simulator

import "fmt"

// GrowthStrategy defines how a list chooses its new capacity on resize.
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

// OneAndAHalfGrowth implements 1.5× growth (used by Java ArrayList, C++ MSVC vector, C# List).
type OneAndAHalfGrowth struct{}

func (OneAndAHalfGrowth) NextCapacity(needed int) int {
	grown := needed + (needed+1)/2
	return max(4, grown)
}

// NoGrowth is a fixed-capacity strategy.
// Returns -1 when capacity is exceeded, signaling overflow.
type NoGrowth struct {
	Cap int
}

func (ng NoGrowth) NextCapacity(needed int) int {
	if needed <= ng.Cap {
		return ng.Cap
	}
	return -1
}

// StrategyByName maps a wire name to a GrowthStrategy.
func StrategyByName(name string) (GrowthStrategy, error) {
	switch name {
	case "cpython":
		return CPythonGrowth{}, nil
	case "doubling":
		return DoublingGrowth{}, nil
	case "1.5x":
		return OneAndAHalfGrowth{}, nil
	case "nogrowth":
		return NoGrowth{Cap: 4}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %q", name)
	}
}
