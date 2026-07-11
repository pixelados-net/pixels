package config

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	currencyconfig "github.com/niflaot/pixels/internal/realm/inventory/currency"
	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/postgres"
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
		I18N,
		Currency,
		Chat,
		RoomEntry,
		RoomModeration,
		Postgres,
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

// I18N extracts translation settings from composed configuration.
func I18N(config AppConfig) i18n.Config {
	return config.I18N
}

// Currency extracts inventory currency settings from composed configuration.
func Currency(config AppConfig) currencyconfig.Config {
	return config.Currency
}

// Chat extracts protocol chat settings from composed configuration.
func Chat(config AppConfig) chatconfig.Config { return config.Chat }

// RoomEntry extracts closed-room entry settings from composed configuration.
func RoomEntry(config AppConfig) roomentry.Config {
	return config.RoomEntry
}

// RoomModeration extracts room moderation settings from composed configuration.
func RoomModeration(config AppConfig) roommoderation.Config {
	return config.RoomModeration
}

// Postgres extracts PostgreSQL settings from composed configuration.
func Postgres(config AppConfig) postgres.Config {
	return config.Postgres
}

// Redis extracts Redis settings from composed configuration.
func Redis(config AppConfig) redis.Config {
	return config.Redis
}

// SSO extracts single sign-on settings from composed configuration.
func SSO(config AppConfig) sso.Config {
	return config.SSO
}
