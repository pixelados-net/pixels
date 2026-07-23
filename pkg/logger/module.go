package logger

import "go.uber.org/fx"

// Module provides zap logging to an Fx dependency graph.
var Module = fx.Module(
	"logger",
	fx.Provide(New),
)
