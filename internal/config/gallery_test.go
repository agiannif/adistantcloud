package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
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
name = "test1"
width = 2560
height = 1707
alt = "Test image 1"

[[rows]]
layout = "half"
[[rows.images]]
name = "test2"
width = 2560
height = 1707
alt = "Test image 2"
[[rows.images]]
name = "test3"
width = 2560
height = 1707
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

	t.Run("reads image name, alt text and dimensions", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-gallery.toml")

		configContent := `
[[rows]]
layout = "full"
[[rows.images]]
name = "test1"
alt = "Test image 1"
width = 2560
height = 1707
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		config, err := ReadGalleryConfig(configPath)
		if err != nil {
			t.Fatalf("ReadGalleryConfig() error = %v", err)
		}

		want := ImageConfig{Name: "test1", Alt: "Test image 1", Width: 2560, Height: 1707}
		if got := config.Rows[0].Images[0]; got != want {
			t.Errorf("Rows[0].Images[0] = %+v, want %+v", got, want)
		}
	})

	t.Run("returns error naming an image without dimensions", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-gallery.toml")

		configContent := `
[[rows]]
layout = "full"
[[rows.images]]
name = "no-dimensions"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadGalleryConfig(configPath)
		if err == nil {
			t.Fatal("ReadGalleryConfig() expected error for image without dimensions, got nil")
		}
		if !strings.Contains(err.Error(), "no-dimensions") {
			t.Errorf("error %q should name the image", err)
		}
	})

	t.Run("returns error for an image without a name", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-gallery.toml")

		configContent := `
[[rows]]
layout = "full"
[[rows.images]]
width = 2560
height = 1707
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		if _, err := ReadGalleryConfig(configPath); err == nil {
			t.Error("ReadGalleryConfig() expected error for image without a name, got nil")
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
name = "img1"
width = 2560
height = 1707
alt = "Image 1"
`

		gallery2 := `
[metadata]
name = "Gallery Two"
short_name = "two"

[[rows]]
layout = "half"
[[rows.images]]
name = "img2"
width = 2560
height = 1707
alt = "Image 2"
[[rows.images]]
name = "img3"
width = 2560
height = 1707
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
name = "img"
width = 2560
height = 1707
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

func TestGalleryConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"land-gallery.toml", "street-gallery.toml", "home.toml", "server.toml"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "old-gallery"), 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GalleryConfigFiles(tmpDir)
	if err != nil {
		t.Fatalf("GalleryConfigFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(tmpDir, "land-gallery.toml"),
		filepath.Join(tmpDir, "street-gallery.toml"),
	}
	if !slices.Equal(got, want) {
		t.Errorf("GalleryConfigFiles() = %v, want %v", got, want)
	}
}

func TestGalleryConfigFilesErrorsForMissingDirectory(t *testing.T) {
	if _, err := GalleryConfigFiles("/nonexistent/directory"); err == nil {
		t.Error("GalleryConfigFiles() expected error for non-existent directory, got nil")
	}
}
