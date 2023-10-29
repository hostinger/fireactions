package freecpu

import (
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler/scorer"
)

// Scorer is a Scorer that scores Nodes based on their free CPU.
type Scorer struct {
	Multiplier float64
}

var _ scorer.Scorer = &Scorer{}

// Name returns the name of the Scorer.
func (s *Scorer) Name() string {
	return "free-cpu"
}

// Score returns the score of the Node.
func (s *Scorer) Score(node *fireactions.Node) (float64, error) {
	return float64(node.CPU.Available()), nil
}

// New returns a new Scorer.
func New() *Scorer {
	return &Scorer{}
}
