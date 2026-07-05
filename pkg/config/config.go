// Package config composes application configuration from focused holders.
package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/redis"
)

// AppConfig composes startup configuration without owning component settings.
type AppConfig struct {
	// App contains application-level settings.
	App appconfig.Config

	// Logger contains zap logger settings.
	Logger logger.Config

	// Redis contains reusable Redis storage settings.
	Redis redis.Config
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

	redis, err := redis.LoadConfig()
	if err != nil {
		return AppConfig{}, err
	}

	return AppConfig{
		App:    app,
		Logger: log,
		Redis:  redis,
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
