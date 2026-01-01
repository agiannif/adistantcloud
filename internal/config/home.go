package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// HomeConfig represents the home page configuration
type HomeConfig struct {
	HeroImages []string `toml:"hero_images"`
}

// ReadHomeConfig reads and parses the home page TOML config file
func ReadHomeConfig(path string) (*HomeConfig, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("cannot find home config file at location: %v", path)
	}

	var config HomeConfig
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to read home config: %w", err)
	}

	return &config, nil
}
