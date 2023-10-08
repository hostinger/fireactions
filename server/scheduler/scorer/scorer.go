package scorer

import "github.com/hostinger/fireactions/server/structs"

// Scorer is an interface that scores nodes based on certain criteria.
type Scorer interface {
	// Name returns the name of the scorer.
	Name() string

	// Score scores nodes based on certain criteria.
	Score(runner *structs.Runner, node *structs.Node) (float64, error)

	// String returns a string representation of the Scorer.
	String() string
}
