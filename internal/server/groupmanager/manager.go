package groupmanager

import (
	"errors"
	"sync"

	"github.com/hostinger/fireactions/internal/structs"
	"github.com/rs/zerolog"
)

var (
	// ErrGroupExists is returned when a Group with the same name already exists.
	ErrGroupExists = errors.New("group already exists")

	// ErrGroupNotFound is returned when a Group with the given name does not exist.
	ErrGroupNotFound = errors.New("group not found")
)

// GroupManager is a thread-safe manager for Groups.
type GroupManager struct {
	groups map[string]*structs.Group
	mu     sync.RWMutex

	log *zerolog.Logger
}

// New creates a new GroupManager.
func New(log *zerolog.Logger) *GroupManager {
	gm := &GroupManager{
		groups: make(map[string]*structs.Group),
		mu:     sync.RWMutex{},
	}

	logger := log.With().Str("component", "groups-manager").Logger()
	gm.log = &logger

	return gm
}

// AddGroups adds multiple Groups to the manager.
func (gm *GroupManager) AddGroups(g ...*structs.Group) error {
	for _, group := range g {
		err := gm.AddGroup(group)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddGroup adds a new Group to the manager.
func (gm *GroupManager) AddGroup(group *structs.Group) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.groups[group.Name] != nil {
		return ErrGroupExists
	}

	gm.groups[group.Name] = group
	gm.log.Info().Msgf("added group: %s", group)
	return nil
}

// GetGroup returns a Group by name.
func (gm *GroupManager) GetGroup(name string) (*structs.Group, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group := gm.groups[name]
	if group == nil {
		return nil, ErrGroupNotFound
	}

	return group, nil
}

// ListGroups returns a list of all Groups.
func (gm *GroupManager) ListGroups() ([]*structs.Group, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	groups := make([]*structs.Group, 0, len(gm.groups))
	for _, group := range gm.groups {
		groups = append(groups, group)
	}

	return groups, nil
}

// DisableGroup disables a Group by name.
func (gm *GroupManager) DisableGroup(name string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group := gm.groups[name]
	if group == nil {
		return ErrGroupNotFound
	}

	group.Enabled = false
	gm.log.Info().Msgf("disabled group: %s", group)
	return nil
}

// EnableGroup enables a Group by name.
func (gm *GroupManager) EnableGroup(name string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group := gm.groups[name]
	if group == nil {
		return ErrGroupNotFound
	}

	group.Enabled = true
	gm.log.Info().Msgf("enabled group: %s", group)
	return nil
}
