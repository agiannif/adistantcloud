package config

import (
	"fmt"
	"os"

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
	Description string
	Images      []ImageConfig
}

type GalleryConfig struct {
	Rows []RowConfig
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
