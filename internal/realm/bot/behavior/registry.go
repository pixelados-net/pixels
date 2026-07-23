// Package behavior owns built-in bot behavior registration.
package behavior

import (
	"errors"
	"strings"
	"sync"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

var (
	// ErrAlreadyRegistered reports a duplicate behavior discriminator.
	ErrAlreadyRegistered = errors.New("bot behavior already registered")
	// ErrInvalidBehavior reports malformed behavior registration.
	ErrInvalidBehavior = errors.New("invalid bot behavior")
)

// Factory creates isolated behavior instances.
type Factory func() sdkbot.Behavior

// Registry stores immutable behavior factories after application startup.
type Registry struct {
	// mutex protects registration and lookup.
	mutex sync.RWMutex
	// factories maps persisted discriminators to constructors.
	factories map[string]Factory
}

// NewRegistry creates an empty behavior registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds one behavior factory.
func (registry *Registry) Register(botType string, factory Factory) error {
	botType = strings.TrimSpace(botType)
	if botType == "" || factory == nil {
		return ErrInvalidBehavior
	}
	instance := factory()
	if instance == nil || instance.Type() != botType {
		return ErrInvalidBehavior
	}
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	if _, exists := registry.factories[botType]; exists {
		return ErrAlreadyRegistered
	}
	registry.factories[botType] = factory
	return nil
}

// For returns one new behavior and safely falls back to generic.
func (registry *Registry) For(botType string) sdkbot.Behavior {
	registry.mutex.RLock()
	factory := registry.factories[botType]
	if factory == nil {
		factory = registry.factories["generic"]
	}
	registry.mutex.RUnlock()
	if factory == nil {
		return nil
	}
	return factory()
}
