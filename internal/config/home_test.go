package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadHomeConfig(t *testing.T) {
	// Create temp directory with test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "home.toml")

	content := `hero_images = [
    "image1.jpg",
    "image2.jpg",
    "image3.jpg"
]`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Read config
	config, err := ReadHomeConfig(configPath)
	if err != nil {
		t.Fatalf("ReadHomeConfig() error = %v", err)
	}

	// Verify hero images
	expectedImages := []string{"image1.jpg", "image2.jpg", "image3.jpg"}
	if len(config.HeroImages) != len(expectedImages) {
		t.Errorf("Expected %d hero images, got %d", len(expectedImages), len(config.HeroImages))
	}

	for i, img := range expectedImages {
		if config.HeroImages[i] != img {
			t.Errorf("Expected hero_images[%d] = %s, got %s", i, img, config.HeroImages[i])
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

func TestReadHomeConfigMissingFile(t *testing.T) {
	_, err := ReadHomeConfig("/nonexistent/path/home.toml")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}
