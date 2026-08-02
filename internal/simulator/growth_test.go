package simulator

import (
	"fmt"
	"testing"
)

func TestCPythonGrowthSequence(t *testing.T) {
	cases := []struct {
		needed int
		want   int
	}{
		{1, 4},
		{5, 8},
		{9, 16},
		{17, 24},
		{25, 32},
		{33, 40},
		{41, 52},
		{53, 64},
		{65, 76},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("needed=%d", tc.needed), func(t *testing.T) {
			got := CPythonGrowth{}.NextCapacity(tc.needed)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})

	}
}

func TestDoublingGrowthSequence(t *testing.T) {
	cases := []struct {
		needed, want int
	}{
		{1, 4},
		{2, 4},
		{3, 4},
		{4, 4},
		{5, 8},
		{8, 8},
		{9, 16},
		{16, 16},
		{17, 32},
		{33, 64},
		{65, 128},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("needed=%d", tc.needed), func(t *testing.T) {
			got := DoublingGrowth{}.NextCapacity(tc.needed)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOneAndAHalfGrowthSequence(t *testing.T) {
	cases := []struct {
		needed, want int
	}{
		{1, 4},
		{2, 4},
		{3, 5},
		{4, 6},
		{5, 8},
		{6, 9},
		{7, 11},
		{8, 12},
		{16, 24},
		{33, 50},
		{64, 96},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("needed=%d", tc.needed), func(t *testing.T) {
			got := OneAndAHalfGrowth{}.NextCapacity(tc.needed)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNoGrowthFits(t *testing.T) {
	cases := []struct {
		cap, needed, want int
	}{
		{4, 1, 4},
		{4, 4, 4},
		{10, 7, 10},
		{10, 10, 10},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("cap=%d/needed=%d", tc.cap, tc.needed), func(t *testing.T) {
			got := NoGrowth{Cap: tc.cap}.NextCapacity(tc.needed)
			if got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNoGrowthPanicsOnOverflow(t *testing.T) {
	cases := []struct {
		cap, needed int
	}{
		{4, 5},
		{10, 11},
		{1, 2},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("cap=%d/needed=%d", tc.cap, tc.needed), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for cap=%d, needed=%d", tc.cap, tc.needed)
				}
			}()
			NoGrowth{Cap: tc.cap}.NextCapacity(tc.needed)
		})
	}
}
