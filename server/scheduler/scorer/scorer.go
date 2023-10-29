package scorer

import "github.com/hostinger/fireactions"

// Scorer is an interface that scores nodes based on certain criteria.
type Scorer interface {
	// Name returns the name of the scorer.
	Name() string

	// Score scores nodes based on certain criteria.
	Score(node *fireactions.Node) (float64, error)
}
