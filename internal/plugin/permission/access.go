// Package permission implements namespaced dynamic permission declarations.
package permission

import (
	"fmt"
	"strings"

	permissiondomain "github.com/niflaot/pixels/internal/permission"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
)

// Access registers permission nodes owned by one plugin scope.
type Access struct {
	// scope stores the plugin namespace.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped permission registrar.
func NewAccess(scope *pluginruntime.Scope) *Access { return &Access{scope: scope} }

// Register declares one plugin-local permission node.
func (access *Access) Register(node string, description string) error {
	local := strings.TrimSpace(node)
	if local == "" || strings.HasPrefix(local, "plugin.") {
		return pluginruntime.ErrInvalidPermission
	}
	full := permissiondomain.Node(fmt.Sprintf("plugin.%s.%s", access.scope.Name(), local))
	if !full.Concrete() {
		return pluginruntime.ErrInvalidPermission
	}
	return permissiondomain.RegisterPluginNode(full, description)
}
