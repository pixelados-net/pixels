package registry

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	// ErrDuplicateKey reports a duplicate canonical descriptor key.
	ErrDuplicateKey = errors.New("duplicate WIRED descriptor key")
	// ErrDuplicateAlias reports an alias claimed by multiple descriptors.
	ErrDuplicateAlias = errors.New("duplicate WIRED descriptor alias")
	// ErrFrozen reports an attempted mutation after registry finalization.
	ErrFrozen = errors.New("WIRED registry is frozen")
)

// Registry stores immutable canonical WIRED descriptors and aliases.
type Registry struct {
	// mutex protects construction and frozen lookup state.
	mutex sync.RWMutex
	// byKey stores descriptors by canonical key.
	byKey map[string]Descriptor
	// aliases maps imported interaction names to canonical keys.
	aliases map[string]string
	// frozen prevents runtime registration.
	frozen bool
}

// New creates an empty mutable registry.
func New() *Registry {
	return &Registry{byKey: make(map[string]Descriptor), aliases: make(map[string]string)}
}

// Canonical creates, validates, and freezes the audited registry.
func Canonical() (*Registry, error) {
	result := New()
	manifest := append(CanonicalManifest(), CompatibilityManifest()...)
	for _, descriptor := range manifest {
		if err := result.Register(descriptor); err != nil {
			return nil, err
		}
	}
	result.Freeze()
	return result, nil
}

// Register adds one descriptor during startup construction.
func (registry *Registry) Register(descriptor Descriptor) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	if registry.frozen {
		return ErrFrozen
	}
	if descriptor.Key == "" || descriptor.Family < FamilyTrigger || descriptor.Family > FamilyHighscore {
		return fmt.Errorf("invalid WIRED descriptor %q", descriptor.Key)
	}
	if _, exists := registry.byKey[descriptor.Key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateKey, descriptor.Key)
	}
	if _, exists := registry.aliases[descriptor.Key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateAlias, descriptor.Key)
	}
	for _, alias := range descriptor.Aliases {
		if alias == "" || alias == descriptor.Key {
			return fmt.Errorf("invalid WIRED alias %q", alias)
		}
		if _, exists := registry.byKey[alias]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateAlias, alias)
		}
		if _, exists := registry.aliases[alias]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateAlias, alias)
		}
	}
	registry.byKey[descriptor.Key] = descriptor
	for _, alias := range descriptor.Aliases {
		registry.aliases[alias] = descriptor.Key
	}
	return nil
}

// Freeze prevents further registration.
func (registry *Registry) Freeze() {
	registry.mutex.Lock()
	registry.frozen = true
	registry.mutex.Unlock()
}

// Resolve returns a descriptor by canonical key or accepted alias.
func (registry *Registry) Resolve(key string) (Descriptor, bool) {
	registry.mutex.RLock()
	canonical, aliased := registry.aliases[key]
	if aliased {
		key = canonical
	}
	descriptor, found := registry.byKey[key]
	registry.mutex.RUnlock()
	return descriptor, found
}

// Manifest returns descriptors in stable key order.
func (registry *Registry) Manifest() []Descriptor {
	registry.mutex.RLock()
	result := make([]Descriptor, 0, len(registry.byKey))
	for _, descriptor := range registry.byKey {
		result = append(result, descriptor)
	}
	registry.mutex.RUnlock()
	sort.Slice(result, func(left int, right int) bool { return result[left].Key < result[right].Key })
	return result
}
