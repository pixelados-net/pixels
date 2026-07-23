package figure

import "github.com/caarlos0/env/v11"

// Config contains the authoritative avatar figure-data source.
type Config struct {
	// Path identifies the Nitro-compatible FigureData JSON or figuredata XML file.
	Path string `env:"PIXELS_FIGURE_DATA_PATH" envDefault:"legacy/Comet-KeyServers-Edition/config/figuredata.xml"`
}

// LoadConfig loads figure-data settings from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }
