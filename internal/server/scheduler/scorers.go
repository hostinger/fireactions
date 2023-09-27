package scheduler

import (
	"github.com/hostinger/fireactions/internal/structs"
)

// Scorer is an interface that scores nodes based on certain criteria.
type Scorer interface {
	// Name returns the name of the scorer.
	Name() string
	// Score scores nodes based on certain criteria.
	Score(runner *structs.Runner, node *structs.Node) (float64, error)
}

var (
	defaultRamScorerMultiplier = 1.0
)

// FreeRamScorer is a Scorer that scores Nodes based on their free RAM.
type FreeRamScorer struct {
	Multiplier float64
}

// Name returns the name of the Scorer.
func (s *FreeRamScorer) Name() string {
	return "free-ram"
}

// Score returns the score of the Node.
func (s *FreeRamScorer) Score(runner *structs.Runner, node *structs.Node) (float64, error) {
	return float64(node.RAM.Available()) * s.Multiplier, nil
}

var (
	defaultCpuScorerMultiplier = 1.0
)

// FreeCpuScorer is a Scorer that scores Nodes based on their free CPU.
type FreeCpuScorer struct {
	Multiplier float64
}

// Name returns the name of the Scorer.
func (s *FreeCpuScorer) Name() string {
	return "free-cpu"
}

// Score returns the score of the Node.
func (s *FreeCpuScorer) Score(runner *structs.Runner, node *structs.Node) (float64, error) {
	return float64(node.CPU.Available()) * s.Multiplier, nil
}
