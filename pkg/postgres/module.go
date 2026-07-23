package postgres

import "go.uber.org/fx"

// Module provides PostgreSQL infrastructure to an Fx dependency graph.
var Module = fx.Module(
	"postgres",
	fx.Provide(NewPool),
)
