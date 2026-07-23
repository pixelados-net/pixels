// Package route implements isolated HTTP and OpenAPI plugin registration.
package route

import (
	"github.com/gofiber/fiber/v2"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
)

// Access enforces one plugin manifest namespace around HTTP declarations.
type Access struct {
	// routes stores shared isolated route declarations.
	routes *pluginroutes.Registry
	// scope stores plugin identity and health.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped route registrar.
func NewAccess(routes *pluginroutes.Registry, scope *pluginruntime.Scope) *Access {
	return &Access{routes: routes, scope: scope}
}

// Mount registers routes only below the calling plugin's namespace.
func (access *Access) Mount(pluginName string, register func(fiber.Router)) error {
	if pluginName != access.scope.Name() {
		return pluginruntime.ErrWrongNamespace
	}
	return access.routes.Mount(pluginName, func(router fiber.Router) {
		defer func() {
			if recovered := recover(); recovered != nil {
				access.scope.Disable()
				panic(recovered)
			}
		}()
		router.Use(func(ctx *fiber.Ctx) (err error) {
			if !access.scope.Enabled() {
				return fiber.NewError(fiber.StatusServiceUnavailable, pluginruntime.ErrPluginDisabled.Error())
			}
			defer func() {
				if recovered := recover(); recovered != nil {
					access.scope.Disable()
					panic(recovered)
				}
			}()
			return ctx.Next()
		})
		register(router)
	})
}

// Describe publishes an independent OpenAPI document for this plugin.
func (access *Access) Describe(pluginName string, spec []byte) error {
	if pluginName != access.scope.Name() {
		return pluginruntime.ErrWrongNamespace
	}
	return access.routes.Describe(pluginName, spec)
}
