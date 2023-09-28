package flavormanager

import (
	"fmt"
	"sync"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
)

var (
	defaultFlavor = &structs.Flavor{Name: "default", VCPUs: 1, MemorySizeMB: 1024, DiskSizeGB: 50}
)

var (
	ErrFlavorNotFound = fmt.Errorf("flavor not found")
	ErrFlavorExists   = fmt.Errorf("flavor already exists")
)

// FlavorManager manages flavors.
type FlavorManager struct {
	flavors       map[string]*structs.Flavor
	defaultFlavor *structs.Flavor
	mu            sync.RWMutex

	log *zerolog.Logger
}

// FlavorManagerOpt is a function that configures a FlavorManager.
type FlavorManagerOpt func(*FlavorManager)

// WithDefaultFlavor sets the default Flavor for the FlavorManager.
func WithDefaultFlavor(flavor *structs.Flavor) FlavorManagerOpt {
	f := func(fm *FlavorManager) {
		fm.defaultFlavor = flavor
	}

	return f
}

// New creates a new FlavorManager.
func New(log *zerolog.Logger, opts ...FlavorManagerOpt) *FlavorManager {
	fm := &FlavorManager{
		flavors: make(map[string]*structs.Flavor),
		mu:      sync.RWMutex{},
		log:     log,
	}

	logger := log.With().Str("component", "flavor-manager").Logger()
	fm.log = &logger

	for _, opt := range opts {
		opt(fm)
	}

	if fm.defaultFlavor == nil {
		fm.defaultFlavor = defaultFlavor
	}

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

// GetFlavor returns a Flavor by name.
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

// GetDefaultFlavor gets the default Flavor.
func (fm *FlavorManager) GetDefaultFlavor() *structs.Flavor {
	return fm.defaultFlavor
}

// SetDefaultFlavor sets the default Flavor.
func (fm *FlavorManager) SetDefaultFlavor(flavor string) error {
	f, err := fm.GetFlavor(flavor)
	if err != nil {
		return err
	}

	fm.defaultFlavor = f
	fm.log.Info().Msgf("set default flavor: %s", f.String())
	return nil
}
