package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Port                  int
	NumRowsPerGalleryPage int `toml:"num_rows_per_gallery_page"`
}

func ReadServerConfig(location string) (*ServerConfig, error) {
	config := ServerConfig{
		// set default values
		Port:                  0,
		NumRowsPerGalleryPage: 3,
	}

	if _, err := os.Stat(location); err != nil {
		slog.Warn("cannot find server config file, using default values")
		return &config, nil
	}

	_, err := toml.DecodeFile(location, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to read server config: %w", err)
	}

	if config.NumRowsPerGalleryPage == 0 {
		slog.Warn("setting number of rows per page in the gallery to zero will result in an empty gallery")
	}

	return &config, nil
}
