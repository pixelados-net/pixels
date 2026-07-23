package http

import (
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	"go.uber.org/fx"
)

// Module provides Fiber HTTP transport to an Fx dependency graph.
var Module = fx.Module(
	"http",
	fx.Provide(pluginroutes.New),
	fx.Provide(NewWithPermissions),
	fx.Invoke(Start),
)
