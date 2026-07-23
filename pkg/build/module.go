package build

import "go.uber.org/fx"

// Module provides build metadata to an Fx dependency graph.
var Module = fx.Module(
	"build",
	fx.Provide(DefaultInfo),
)
