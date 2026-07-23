package redis

import "go.uber.org/fx"

// Module provides Redis storage to an Fx dependency graph.
var Module = fx.Module(
	"redis",
	fx.Provide(New),
)
