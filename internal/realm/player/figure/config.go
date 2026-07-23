package figure

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains the authoritative avatar figure-data source.
type Config struct {
	// URL identifies the Nitro-compatible FigureData JSON or figuredata XML endpoint.
	URL string `env:"PIXELS_FIGURE_DATA_URL" envDefault:"https://storageapi.pixelados.net/assets-prod/gamedata/FigureData.json"`
	// Path identifies a local FigureData JSON or figuredata XML file and overrides URL.
	Path string `env:"PIXELS_FIGURE_DATA_PATH" envDefault:""`
	// Timeout bounds a remote figure-data request.
	Timeout time.Duration `env:"PIXELS_FIGURE_DATA_TIMEOUT" envDefault:"15s"`
	// MaxBytes bounds the downloaded or locally read figure-data document.
	MaxBytes int64 `env:"PIXELS_FIGURE_DATA_MAX_BYTES" envDefault:"16777216"`
}

// LoadConfig loads figure-data settings from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }
