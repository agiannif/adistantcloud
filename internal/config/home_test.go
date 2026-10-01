package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadHomeConfig(t *testing.T) {
	// Create temp directory with test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "home.toml")

	content := `
[[hero_images]]
name = "image1"
width = 2560
height = 1707

[[hero_images]]
name = "image2"
width = 1707
height = 2560

[[hero_images]]
name = "image3"
width = 2560
height = 1707
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read config
	config, err := ReadHomeConfig(configPath)
	if err != nil {
		t.Fatalf("ReadHomeConfig() error = %v", err)
	}

	// Verify hero images
	expectedImages := []ImageConfig{
		{Name: "image1", Width: 2560, Height: 1707},
		{Name: "image2", Width: 1707, Height: 2560},
		{Name: "image3", Width: 2560, Height: 1707},
	}
	if len(config.HeroImages) != len(expectedImages) {
		t.Fatalf("Expected %d hero images, got %d", len(expectedImages), len(config.HeroImages))
	}

	for i, img := range expectedImages {
		if config.HeroImages[i] != img {
			t.Errorf("Expected hero_images[%d] = %+v, got %+v", i, img, config.HeroImages[i])
		}
	}
}

func TestReadHomeConfigEmptyImages(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "home.toml")

	content := `hero_images = []`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config, err := ReadHomeConfig(configPath)
	if err != nil {
		t.Fatalf("ReadHomeConfig() error = %v", err)
	}

	if len(config.HeroImages) != 0 {
		t.Errorf("Expected 0 hero images for empty array, got %d", len(config.HeroImages))
	}
}

func TestReadHomeConfigImageWithoutDimensions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "home.toml")

	content := `
[[hero_images]]
name = "no-dimensions"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadHomeConfig(configPath)
	if err == nil {
		t.Fatal("Expected error for hero image without dimensions, got nil")
	}
	if !strings.Contains(err.Error(), "no-dimensions") {
		t.Errorf("error %q should name the image", err)
	}
}

func TestReadHomeConfigMissingFile(t *testing.T) {
	_, err := ReadHomeConfig("/nonexistent/path/home.toml")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}
