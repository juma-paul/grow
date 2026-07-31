package simulator

type GrowthStrategy interface {
	NextCapacity(needed int) int
}