package config

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	currencyconfig "github.com/niflaot/pixels/internal/realm/inventory/currency"
	realmmessenger "github.com/niflaot/pixels/internal/realm/messenger"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	realmsubscription "github.com/niflaot/pixels/internal/realm/subscription"
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
		Plugin,
		Chat,
		Messenger,
		Subscription,
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

// Plugin extracts native plugin runtime settings from composed configuration.
func Plugin(config AppConfig) pluginconfig.Config {
	return config.Plugin
}

// Chat extracts protocol chat settings from composed configuration.
func Chat(config AppConfig) chatconfig.Config { return config.Chat }

// Messenger extracts social communication configuration.
func Messenger(config AppConfig) realmmessenger.Config { return config.Messenger }

// Subscription extracts club scheduler and reward configuration.
func Subscription(config AppConfig) realmsubscription.Config { return config.Subscription }

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
