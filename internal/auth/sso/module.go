package sso

import "go.uber.org/fx"

// Module provides SSO authentication services.
var Module = fx.Module(
	"sso",
	fx.Provide(New),
)
