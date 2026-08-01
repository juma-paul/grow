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
