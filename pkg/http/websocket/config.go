// Package websocket adapts Fiber WebSockets to pixel-protocol sessions.
package websocket

import (
	"errors"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
)

var (
	// ErrQueueFull reports outbound WebSocket backpressure.
	ErrQueueFull = errors.New("websocket queue full")
)

// Module provides WebSocket transport dependencies.
var Module = fx.Module("http-websocket", fx.Provide(LoadConfig, New))

// Config contains WebSocket transport settings.
type Config struct {
	// QueueSize is the bounded outbound packet queue length.
	QueueSize int `env:"PIXELS_WS_QUEUE_SIZE" envDefault:"256"`
	// WriteTimeout limits one WebSocket write.
	WriteTimeout time.Duration `env:"PIXELS_WS_WRITE_TIMEOUT" envDefault:"5s"`
	// ReadTimeout limits one WebSocket read.
	ReadTimeout time.Duration `env:"PIXELS_WS_READ_TIMEOUT" envDefault:"75s"`
	// PingInterval controls server heartbeat pings.
	PingInterval time.Duration `env:"PIXELS_WS_PING_INTERVAL" envDefault:"30s"`
	// PongTimeout controls idle disconnects.
	PongTimeout time.Duration `env:"PIXELS_WS_PONG_TIMEOUT" envDefault:"60s"`
	// CloseGrace limits graceful close flushing.
	CloseGrace time.Duration `env:"PIXELS_WS_CLOSE_GRACE" envDefault:"2s"`
}

// LoadConfig reads WebSocket configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills defensive defaults for manually created config values.
func (config Config) Normalize() Config {
	if config.QueueSize <= 0 {
		config.QueueSize = 256
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 5 * time.Second
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 75 * time.Second
	}
	if config.PingInterval <= 0 {
		config.PingInterval = 30 * time.Second
	}
	if config.PongTimeout <= 0 {
		config.PongTimeout = 60 * time.Second
	}
	if config.CloseGrace <= 0 {
		config.CloseGrace = 2 * time.Second
	}

	return config
}
