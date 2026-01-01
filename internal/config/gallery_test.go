package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadGalleryConfig(t *testing.T) {
	t.Run("reads valid gallery config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-gallery.toml")

		configContent := `
[metadata]
name = "Test Gallery"
short_name = "test"

[[rows]]
layout = "full"
[[rows.images]]
file = "test1.jpg"
alt = "Test image 1"

[[rows]]
layout = "half"
[[rows.images]]
file = "test2.jpg"
alt = "Test image 2"
[[rows.images]]
file = "test3.jpg"
alt = "Test image 3"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		config, err := ReadGalleryConfig(configPath)
		if err != nil {
			t.Fatalf("ReadGalleryConfig() error = %v", err)
		}

		if config.Metadata.Name != "Test Gallery" {
			t.Errorf("Metadata.Name = %q, want %q", config.Metadata.Name, "Test Gallery")
		}

		if config.Metadata.ShortName != "test" {
			t.Errorf("Metadata.ShortName = %q, want %q", config.Metadata.ShortName, "test")
		}

		if len(config.Rows) != 2 {
			t.Fatalf("len(Rows) = %d, want 2", len(config.Rows))
		}

		if config.Rows[0].Layout != LayoutFull {
			t.Errorf("Rows[0].Layout = %q, want %q", config.Rows[0].Layout, LayoutFull)
		}

		if len(config.Rows[0].Images) != 1 {
			t.Fatalf("len(Rows[0].Images) = %d, want 1", len(config.Rows[0].Images))
		}

		if config.Rows[1].Layout != LayoutHalf {
			t.Errorf("Rows[1].Layout = %q, want %q", config.Rows[1].Layout, LayoutHalf)
		}

		if len(config.Rows[1].Images) != 2 {
			t.Fatalf("len(Rows[1].Images) = %d, want 2", len(config.Rows[1].Images))
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := ReadGalleryConfig("/nonexistent/path/config.toml")
		if err == nil {
			t.Error("ReadGalleryConfig() expected error for non-existent file, got nil")
		}
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.toml")

		invalidContent := `this is not valid TOML [[[`
		if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadGalleryConfig(configPath)
		if err == nil {
			t.Error("ReadGalleryConfig() expected error for invalid TOML, got nil")
		}
	})
}

func TestReadGalleryConfigsIn(t *testing.T) {
	t.Run("reads multiple gallery configs from directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		gallery1 := `
[metadata]
name = "Gallery One"
short_name = "one"

[[rows]]
layout = "full"
[[rows.images]]
file = "img1.jpg"
alt = "Image 1"
`

		gallery2 := `
[metadata]
name = "Gallery Two"
short_name = "two"

[[rows]]
layout = "half"
[[rows.images]]
file = "img2.jpg"
alt = "Image 2"
[[rows.images]]
file = "img3.jpg"
alt = "Image 3"
`

		if err := os.WriteFile(filepath.Join(tmpDir, "gallery-one.toml"), []byte(gallery1), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(tmpDir, "gallery-two.toml"), []byte(gallery2), 0644); err != nil {
			t.Fatal(err)
		}

		configs, err := ReadGalleryConfigsIn(tmpDir)
		if err != nil {
			t.Fatalf("ReadGalleryConfigsIn() error = %v", err)
		}

		if len(configs) != 2 {
			t.Fatalf("len(configs) = %d, want 2", len(configs))
		}

		if _, exists := configs["one"]; !exists {
			t.Error("configs missing key 'one'")
		}

		if _, exists := configs["two"]; !exists {
			t.Error("configs missing key 'two'")
		}

		if configs["one"].Metadata.Name != "Gallery One" {
			t.Errorf("configs['one'].Metadata.Name = %q, want %q", configs["one"].Metadata.Name, "Gallery One")
		}

		if configs["two"].Metadata.Name != "Gallery Two" {
			t.Errorf("configs['two'].Metadata.Name = %q, want %q", configs["two"].Metadata.Name, "Gallery Two")
		}
	})

	t.Run("ignores non-gallery files", func(t *testing.T) {
		tmpDir := t.TempDir()

		galleryConfig := `
[metadata]
name = "Gallery"
short_name = "gallery"

[[rows]]
layout = "full"
[[rows.images]]
file = "img.jpg"
alt = "Image"
`

		if err := os.WriteFile(filepath.Join(tmpDir, "gallery.toml"), []byte(galleryConfig), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(tmpDir, "server.toml"), []byte("port = 8080"), 0644); err != nil {
			t.Fatal(err)
		}

		configs, err := ReadGalleryConfigsIn(tmpDir)
		if err != nil {
			t.Fatalf("ReadGalleryConfigsIn() error = %v", err)
		}

		if len(configs) != 1 {
			t.Fatalf("len(configs) = %d, want 1 (should ignore server.toml)", len(configs))
		}

		if _, exists := configs["gallery"]; !exists {
			t.Error("configs missing key 'gallery'")
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, err := ReadGalleryConfigsIn("/nonexistent/directory")
		if err == nil {
			t.Error("ReadGalleryConfigsIn() expected error for non-existent directory, got nil")
		}
	})
}
