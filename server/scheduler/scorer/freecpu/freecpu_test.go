package freecpu

import (
	"fmt"
	"testing"

	"github.com/hostinger/fireactions/server/structs"
	"github.com/stretchr/testify/assert"
)

func TestScorer(t *testing.T) {
	testCases := []struct {
		name       string
		runner     *structs.Runner
		node       *structs.Node
		multiplier float64
		want       float64
	}{
		{
			name: "score is correct with multiplier 1.0",
			node: &structs.Node{CPU: structs.Resource{Capacity: 1000, OvercommitRatio: 1.0}},
			runner: &structs.Runner{
				Flavor: &structs.Flavor{VCPUs: 1},
			},
			want:       1000,
			multiplier: 1.0,
		},
		{
			name: "score is correct with multiplier 2.0",
			node: &structs.Node{CPU: structs.Resource{Capacity: 1000, OvercommitRatio: 1.0}},
			runner: &structs.Runner{
				Flavor: &structs.Flavor{VCPUs: 1},
			},
			want:       2000,
			multiplier: 2.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scorer := New(tc.multiplier)
			got, _ := scorer.Score(tc.runner, tc.node)
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}

			assert.Equal(t, "free-cpu", scorer.Name())
			assert.Equal(t, fmt.Sprintf("free-cpu (Multiplier: %.2f)", tc.multiplier), scorer.String())
		})
	}
}
