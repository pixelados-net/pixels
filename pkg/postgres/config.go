// Package postgres contains reusable PostgreSQL infrastructure.
package postgres

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains PostgreSQL connection settings.
type Config struct {
	// Host is the PostgreSQL server host.
	Host string `env:"PIXELS_POSTGRES_HOST" envDefault:"localhost"`

	// Port is the PostgreSQL server port.
	Port int `env:"PIXELS_POSTGRES_PORT" envDefault:"5432"`

	// Database is the PostgreSQL database name.
	Database string `env:"PIXELS_POSTGRES_DATABASE" envDefault:"pixels"`

	// User is the PostgreSQL username.
	User string `env:"PIXELS_POSTGRES_USER" envDefault:"pixels"`

	// Password is the PostgreSQL password.
	Password string `env:"PIXELS_POSTGRES_PASSWORD" envDefault:"pixels"`

	// SSLMode is the PostgreSQL TLS mode.
	SSLMode string `env:"PIXELS_POSTGRES_SSL_MODE" envDefault:"disable"`

	// MaxConns is the maximum pool connection count.
	MaxConns int32 `env:"PIXELS_POSTGRES_MAX_CONNS" envDefault:"10"`

	// MinConns is the minimum pool connection count.
	MinConns int32 `env:"PIXELS_POSTGRES_MIN_CONNS" envDefault:"1"`

	// ConnectTimeout is the maximum duration for connection creation.
	ConnectTimeout time.Duration `env:"PIXELS_POSTGRES_CONNECT_TIMEOUT" envDefault:"5s"`

	// StatementTimeout is the maximum duration expected for one statement.
	StatementTimeout time.Duration `env:"PIXELS_POSTGRES_STATEMENT_TIMEOUT" envDefault:"5s"`

	// HealthTimeout is the maximum duration for health pings.
	HealthTimeout time.Duration `env:"PIXELS_POSTGRES_HEALTH_TIMEOUT" envDefault:"2s"`
}

// LoadConfig reads PostgreSQL configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// DSN returns the PostgreSQL connection string.
func (config Config) DSN() string {
	values := config.query()
	values.Set("password", config.Password)

	return config.url(values)
}

// MaskedDSN returns a PostgreSQL connection string without exposing secrets.
func (config Config) MaskedDSN() string {
	values := config.query()
	values.Set("password", "xxxxx")

	return config.url(values)
}

// query returns PostgreSQL query settings.
func (config Config) query() url.Values {
	values := url.Values{}
	values.Set("sslmode", config.SSLMode)
	values.Set("connect_timeout", strconv.Itoa(int(config.ConnectTimeout.Seconds())))

	return values
}

// url formats the PostgreSQL connection string.
func (config Config) url(values url.Values) string {
	location := url.URL{
		Scheme:   "postgres",
		User:     url.User(config.User),
		Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Path:     config.Database,
		RawQuery: values.Encode(),
	}

	return location.String()
}
