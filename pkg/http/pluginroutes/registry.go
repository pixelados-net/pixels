// Package pluginroutes stores isolated HTTP routes registered by dynamic plugins.
package pluginroutes

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	// ErrInvalidPlugin reports an unsafe or empty plugin namespace.
	ErrInvalidPlugin = errors.New("invalid plugin namespace")
	// ErrInvalidRegistration reports a nil route callback or invalid document.
	ErrInvalidRegistration = errors.New("invalid plugin route registration")
	// ErrAlreadyRegistered reports duplicate plugin routes or documentation.
	ErrAlreadyRegistered = errors.New("plugin route already registered")
)

// registration stores one plugin's route callback and optional OpenAPI document.
type registration struct {
	// mount installs handlers below the plugin group.
	mount func(fiber.Router)
	// spec stores one validated independent OpenAPI document.
	spec []byte
}

// Registry stores route declarations until the Fiber application is assembled.
type Registry struct {
	// mutex protects route declarations during startup.
	mutex sync.RWMutex
	// registrations stores declarations by plugin namespace.
	registrations map[string]registration
}

// New creates an empty plugin route registry.
func New() *Registry {
	return &Registry{registrations: make(map[string]registration)}
}

// Mount stores one plugin route callback.
func (registry *Registry) Mount(pluginName string, mount func(fiber.Router)) error {
	name := strings.TrimSpace(pluginName)
	if !validName(name) {
		return fmt.Errorf("%w: %q", ErrInvalidPlugin, pluginName)
	}
	if mount == nil {
		return ErrInvalidRegistration
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	entry := registry.registrations[name]
	if entry.mount != nil {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, name)
	}
	entry.mount = mount
	registry.registrations[name] = entry

	return nil
}

// Describe stores one plugin-owned OpenAPI document.
func (registry *Registry) Describe(pluginName string, spec []byte) error {
	name := strings.TrimSpace(pluginName)
	if !validName(name) {
		return fmt.Errorf("%w: %q", ErrInvalidPlugin, pluginName)
	}
	if !json.Valid(spec) {
		return ErrInvalidRegistration
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	entry := registry.registrations[name]
	if len(entry.spec) != 0 {
		return fmt.Errorf("%w: %s openapi", ErrAlreadyRegistered, name)
	}
	entry.spec = append([]byte(nil), spec...)
	registry.registrations[name] = entry

	return nil
}

// Register installs every stored plugin group before the authenticated fallback.
func (registry *Registry) Register(app *fiber.App, log *zap.Logger) {
	for _, name := range registry.names() {
		entry := registry.entry(name)
		group := app.Group("/plugins/" + name)
		group.Use(recoverMiddleware(name, log))
		if len(entry.spec) != 0 {
			spec := append([]byte(nil), entry.spec...)
			group.Get("/openapi.json", func(ctx *fiber.Ctx) error {
				ctx.Type("json")
				return ctx.Send(spec)
			})
		}
		mountSafely(name, entry.mount, group, log)
	}
}

// names returns plugin namespaces in deterministic order.
func (registry *Registry) names() []string {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()
	names := make([]string, 0, len(registry.registrations))
	for name := range registry.registrations {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// entry returns a stable registration snapshot.
func (registry *Registry) entry(name string) registration {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()
	entry := registry.registrations[name]
	entry.spec = append([]byte(nil), entry.spec...)

	return entry
}

// mountSafely prevents a plugin mount panic from aborting host startup.
func mountSafely(name string, mount func(fiber.Router), router fiber.Router, log *zap.Logger) {
	if mount == nil {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil && log != nil {
			log.Error("plugin route mount panicked", zap.String("plugin", name), zap.Any("panic", recovered))
		}
	}()
	mount(router)
}

// recoverMiddleware converts plugin handler panics into isolated HTTP errors.
func recoverMiddleware(name string, log *zap.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if log != nil {
					log.Error("plugin route panicked", zap.String("plugin", name), zap.Any("panic", recovered))
				}
				err = fiber.NewError(fiber.StatusInternalServerError, "plugin route failed")
			}
		}()
		return ctx.Next()
	}
}

// validName reports whether a namespace is safe inside one URL path segment.
func validName(name string) bool {
	if name == "" || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for _, value := range name {
		if (value < 'a' || value > 'z') && (value < '0' || value > '9') && value != '-' {
			return false
		}
	}

	return true
}
