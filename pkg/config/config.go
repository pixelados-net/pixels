// Package config composes application configuration from focused holders.
package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
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
)

// AppConfig composes startup configuration without owning component settings.
type AppConfig struct {
	// App contains application-level settings.
	App appconfig.Config

	// Logger contains zap logger settings.
	Logger logger.Config

	// I18N contains translation catalog settings.
	I18N i18n.Config

	// Currency contains inventory currency settings.
	Currency currencyconfig.Config

	// Chat contains protocol chat limits and history settings.
	Chat chatconfig.Config

	// RoomEntry contains closed-room entry settings.
	RoomEntry roomentry.Config

	// RoomModeration contains room moderation duration limits.
	RoomModeration roommoderation.Config

	// Postgres contains PostgreSQL storage settings.
	Postgres postgres.Config

	// Redis contains reusable Redis storage settings.
	Redis redis.Config

	// SSO contains single sign-on ticket settings.
	SSO sso.Config
}

// Load reads dotenv files and composes all configuration holders.
func Load(paths ...string) (AppConfig, error) {
	if err := loadDotenv(paths); err != nil {
		return AppConfig{}, err
	}

	app, err := appconfig.Load()
	if err != nil {
		return AppConfig{}, err
	}

	log, err := logger.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	translations, err := i18n.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	currency, err := currencyconfig.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	chat, err := chatconfig.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	roomEntry, err := roomentry.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	roomModeration, err := roommoderation.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	postgres, err := postgres.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	redis, err := redis.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	sso, err := sso.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	return AppConfig{
		App:            app,
		Logger:         log,
		I18N:           translations,
		Currency:       currency,
		Chat:           chat,
		RoomEntry:      roomEntry,
		RoomModeration: roomModeration,
		Postgres:       postgres,
		Redis:          redis,
		SSO:            sso,
	}, nil
}

// loadDotenv loads explicit dotenv files or an optional local dotenv file.
func loadDotenv(paths []string) error {
	if len(paths) > 0 {
		return godotenv.Load(paths...)
	}

	if err := godotenv.Load(); err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) && errors.Is(pathError.Err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return nil
}
