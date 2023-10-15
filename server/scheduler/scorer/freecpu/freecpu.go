package freecpu

import (
	"fmt"

	"github.com/hostinger/fireactions/server/models"
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
func (s *Scorer) Score(runner *models.Runner, node *models.Node) (float64, error) {
	return float64(node.CPU.Available()) * s.Multiplier, nil
}

// String returns a string representation of the Scorer.
func (s *Scorer) String() string {
	return fmt.Sprintf("%s (Multiplier: %.2f)", s.Name(), s.Multiplier)
}

// New returns a new Scorer.
func New(multiplier float64) *Scorer {
	return &Scorer{
		Multiplier: multiplier,
	}
}
