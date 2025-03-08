package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Layout string

const (
	LayoutFull      Layout = "full"
	LayoutHalf      Layout = "half"
	LayoutSplit     Layout = "split"
	LayoutFlipSplit Layout = "flip split"
	LayoutCenter    Layout = "center"
	LayoutSection   Layout = "section"
)

type ImageConfig struct {
	File string
	Alt  string
}

type RowConfig struct {
	Layout      Layout
	Title       string
	ShortTitle  string `toml:"short_title"`
	Description string
	Images      []ImageConfig
}

type GalleryMetadata struct {
	Name      string
	ShortName string `toml:"short_name"`
}

type GalleryConfig struct {
	Rows     []RowConfig
	Metadata GalleryMetadata
}

func ReadGalleryConfig(location string) (*GalleryConfig, error) {
	if _, err := os.Stat(location); err != nil {
		return nil, fmt.Errorf("cannot find config file at location: %v", location)
	}

	var gallery GalleryConfig
	_, err := toml.DecodeFile(location, &gallery)
	if err != nil {
		return nil, fmt.Errorf("failed to read gallery config: %w", err)
	}

	return &gallery, nil
}

func ReadGalleryConfigsIn(dir string) (map[string]*GalleryConfig, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read gallery directory %s: %w", dir, err)
	}

	configs := make(map[string]*GalleryConfig)
	for _, file := range files {
		if !file.IsDir() && strings.Contains(strings.ToLower(file.Name()), "gallery") {
			filePath := filepath.Join(dir, file.Name())
			config, err := ReadGalleryConfig(filePath)
			if err != nil {
				return nil, err
			}
			configs[config.Metadata.ShortName] = config
		}
	}

	return configs, nil
}
