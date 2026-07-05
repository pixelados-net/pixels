package config

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
)

// Module provides application configuration to an Fx dependency graph.
var Module = fx.Module(
	"config",
	fx.Provide(
		New,
		App,
		Logger,
		Redis,
		SSO,
	),
)

// New loads application configuration for dependency injection.
func New() (AppConfig, error) {
	return Load()
}

// App extracts application-level settings from composed configuration.
func App(config AppConfig) appconfig.Config {
	return config.App
}

// Logger extracts logger settings from composed configuration.
func Logger(config AppConfig) logger.Config {
	return config.Logger
}

// Redis extracts Redis settings from composed configuration.
func Redis(config AppConfig) redis.Config {
	return config.Redis
}

// SSO extracts single sign-on settings from composed configuration.
func SSO(config AppConfig) sso.Config {
	return config.SSO
}
