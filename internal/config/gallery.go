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

// ImageConfig describes a photograph by the name its AVIF variants are derived
// from. Width and Height are the pixel dimensions of the exported original and
// are recorded by the image pipeline.
type ImageConfig struct {
	Name   string
	Alt    string
	Width  int
	Height int
}

func (i ImageConfig) validate() error {
	if i.Name == "" {
		return fmt.Errorf("image has no name")
	}
	if i.Width <= 0 || i.Height <= 0 {
		return fmt.Errorf("image %q has no width and height, run make photos", i.Name)
	}
	return nil
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

	for _, row := range gallery.Rows {
		for _, image := range row.Images {
			if err := image.validate(); err != nil {
				return nil, fmt.Errorf("invalid gallery config %s: %w", location, err)
			}
		}
	}

	return &gallery, nil
}

// GalleryConfigFiles returns the paths of the gallery config files in dir.
func GalleryConfigFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read gallery directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "gallery") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

func ReadGalleryConfigsIn(dir string) (map[string]*GalleryConfig, error) {
	files, err := GalleryConfigFiles(dir)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]*GalleryConfig)
	for _, file := range files {
		config, err := ReadGalleryConfig(file)
		if err != nil {
			return nil, err
		}
		configs[config.Metadata.ShortName] = config
	}

	return configs, nil
}
