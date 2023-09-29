package flavormanager

import (
	"fmt"
	"sync"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
)

var (
	// ErrFlavorNotFound is returned when a Flavor with the given name does not exist.
	ErrFlavorNotFound = fmt.Errorf("flavor not found")

	// ErrFlavorExists is returned when a Flavor with the same name already exists.
	ErrFlavorExists = fmt.Errorf("flavor already exists")
)

// FlavorManager is a thread-safe manager for Flavors.
type FlavorManager struct {
	flavors map[string]*structs.Flavor
	mu      sync.RWMutex

	log *zerolog.Logger
}

// New creates a new FlavorManager.
func New(log *zerolog.Logger) *FlavorManager {
	fm := &FlavorManager{
		flavors: make(map[string]*structs.Flavor),
		mu:      sync.RWMutex{},
		log:     log,
	}

	logger := log.With().Str("component", "flavor-manager").Logger()
	fm.log = &logger

	return fm
}

// AddFlavor adds a Flavor to the FlavorManager.
func (fm *FlavorManager) AddFlavor(f *structs.Flavor) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	_, ok := fm.flavors[f.Name]
	if ok {
		return ErrFlavorExists
	}

	fm.flavors[f.Name] = f
	fm.log.Info().Msgf("added flavor: %s", f.String())
	return nil
}

// AddFlavors adds multiple Flavors to the FlavorManager.
func (fm *FlavorManager) AddFlavors(f ...*structs.Flavor) error {
	for _, flavor := range f {
		err := fm.AddFlavor(flavor)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFlavor returns a Flavor by name. In case the Flavor is not found or is disabled, an error is returned.
func (fm *FlavorManager) GetFlavor(name string) (*structs.Flavor, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, f := range fm.flavors {
		if f.Name != name {
			continue
		}

		return f, nil
	}

	return nil, ErrFlavorNotFound
}

// ListFlavors returns a list of all Flavors.
func (fm *FlavorManager) ListFlavors() ([]*structs.Flavor, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	flavors := make([]*structs.Flavor, 0, len(fm.flavors))
	for _, f := range fm.flavors {
		flavors = append(flavors, f)
	}

	return flavors, nil
}

// DisableFlavor disables a Flavor by name.
func (fm *FlavorManager) DisableFlavor(name string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	f, ok := fm.flavors[name]
	if !ok {
		return ErrFlavorNotFound
	}

	f.Enabled = false
	fm.log.Info().Msgf("disabled flavor: %s", f)
	return nil
}

// EnableFlavor enables a Flavor by name.
func (fm *FlavorManager) EnableFlavor(name string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	f, ok := fm.flavors[name]
	if !ok {
		return ErrFlavorNotFound
	}

	f.Enabled = true
	fm.log.Info().Msgf("enabled flavor: %s", f)
	return nil
}
