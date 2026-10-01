package imageprep

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func requireAvifenc(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("avifenc"); err != nil {
		t.Fatal("avifenc is required to run these tests, see the development dependencies in readme.adoc")
	}
}

func gradient(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x * 255 / width), G: uint8(y * 255 / height), B: 128, A: 255})
		}
	}
	return img
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}

func fileNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestResizeKeepsAspectRatio(t *testing.T) {
	tests := []struct {
		name                  string
		srcWidth, srcHeight   int
		width                 int
		wantWidth, wantHeight int
	}{
		{"landscape", 400, 200, 100, 100, 50},
		{"portrait", 300, 600, 150, 150, 300},
		{"rounds to the nearest pixel", 2560, 1707, 640, 640, 427},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resize(gradient(tt.srcWidth, tt.srcHeight), tt.width).Bounds()

			if got.Dx() != tt.wantWidth || got.Dy() != tt.wantHeight {
				t.Errorf("Resize to width %d gave %dx%d, want %dx%d", tt.width, got.Dx(), got.Dy(), tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestEncodeAVIFWritesAnAVIFFile(t *testing.T) {
	requireAvifenc(t)
	path := filepath.Join(t.TempDir(), "out.avif")

	if err := EncodeAVIF(gradient(64, 48), path, 50); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 12 || string(data[4:12]) != "ftypavif" {
		t.Errorf("output does not start with an AVIF file type box: %q", data[:min(len(data), 12)])
	}
}

func TestEncodeAVIFLowerQualityIsSmaller(t *testing.T) {
	requireAvifenc(t)
	dir := t.TempDir()
	img := gradient(256, 256)

	if err := EncodeAVIF(img, filepath.Join(dir, "high.avif"), 90); err != nil {
		t.Fatal(err)
	}
	if err := EncodeAVIF(img, filepath.Join(dir, "low.avif"), 10); err != nil {
		t.Fatal(err)
	}

	high, _ := os.Stat(filepath.Join(dir, "high.avif"))
	low, _ := os.Stat(filepath.Join(dir, "low.avif"))
	if low.Size() >= high.Size() {
		t.Errorf("quality 10 produced %d bytes, quality 90 produced %d bytes; quality is not being applied", low.Size(), high.Size())
	}
}

func TestEncodeAVIFErrorsWhenOutputDirectoryIsMissing(t *testing.T) {
	requireAvifenc(t)
	path := filepath.Join(t.TempDir(), "missing", "out.avif")

	if err := EncodeAVIF(gradient(64, 48), path, 50); err == nil {
		t.Error("expected an error when the output directory does not exist")
	}
}

func TestEncodeAVIFErrorsClearlyWhenAvifencIsNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := EncodeAVIF(gradient(64, 48), filepath.Join(t.TempDir(), "out.avif"), 50)

	if err == nil || !strings.Contains(err.Error(), "avifenc") {
		t.Errorf("error = %v, want one that mentions avifenc", err)
	}
}

func TestOriginalPathsListsPNGsOnly(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "b.png"), gradient(8, 8))
	writePNG(t, filepath.Join(dir, "a.PNG"), gradient(8, 8))
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "old.jpg"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub.png"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := OriginalPaths(dir)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Join(dir, "a.PNG"), filepath.Join(dir, "b.png")}
	if !slices.Equal(got, want) {
		t.Errorf("OriginalPaths = %v, want %v", got, want)
	}
}

func TestOriginalPathsErrorsWhenThereAreNone(t *testing.T) {
	if _, err := OriginalPaths(t.TempDir()); err == nil {
		t.Error("expected an error for a directory without PNG originals")
	}
}

func TestOriginalPathsErrorsForMissingDirectory(t *testing.T) {
	if _, err := OriginalPaths(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected an error for a missing directory")
	}
}

func TestGenerateVariantsWritesEveryWidthAndReturnsDimensions(t *testing.T) {
	requireAvifenc(t)
	originals, out := t.TempDir(), t.TempDir()
	path := filepath.Join(originals, "wide.png")
	writePNG(t, path, gradient(800, 400))

	name, dimensions, err := GenerateVariants(path, out, 50)
	if err != nil {
		t.Fatal(err)
	}

	if name != "wide" {
		t.Errorf("name = %q, want %q", name, "wide")
	}
	if dimensions != (Dimensions{Width: 800, Height: 400}) {
		t.Errorf("dimensions = %+v, want 800x400", dimensions)
	}
	want := []string{"wide-640.avif", "wide-800.avif"}
	if got := fileNames(t, out); !slices.Equal(got, want) {
		t.Errorf("generated files = %v, want %v", got, want)
	}
}

func TestGenerateVariantsNeverUpscales(t *testing.T) {
	requireAvifenc(t)
	originals, out := t.TempDir(), t.TempDir()
	path := filepath.Join(originals, "small.png")
	writePNG(t, path, gradient(300, 600))

	if _, _, err := GenerateVariants(path, out, 50); err != nil {
		t.Fatal(err)
	}

	if got := fileNames(t, out); !slices.Equal(got, []string{"small-300.avif"}) {
		t.Errorf("generated files = %v, want only small-300.avif", got)
	}
}

func TestGenerateVariantsErrorsNamingAnInvalidPNG(t *testing.T) {
	requireAvifenc(t)
	originals := t.TempDir()
	path := filepath.Join(originals, "broken.png")
	if err := os.WriteFile(path, []byte("not a png"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := GenerateVariants(path, t.TempDir(), 50)

	if err == nil || !strings.Contains(err.Error(), "broken.png") {
		t.Errorf("error = %v, want one naming broken.png", err)
	}
}

func TestUpdateConfigsRecordsDimensionsInGalleryAndHomeConfigs(t *testing.T) {
	dir := t.TempDir()
	gallery := "[metadata]\nname = \"g\"\n\n# comment\n[[rows.images]]\nname = \"larch-1\"\n"
	home := "[[hero_images]]\nname = \"larch-1\"\n"
	server := "port = 1234\n"
	for name, content := range map[string]string{"land-gallery.toml": gallery, "home.toml": home, "server.toml": server} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	unknown, err := UpdateConfigs(dir, map[string]Dimensions{"larch-1": {Width: 2560, Height: 1707}})
	if err != nil {
		t.Fatal(err)
	}

	read := func(name string) string {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if got := read("land-gallery.toml"); got != "[metadata]\nname = \"g\"\n\n# comment\n[[rows.images]]\nname = \"larch-1\"\nwidth = 2560\nheight = 1707\n" {
		t.Errorf("gallery config = %q", got)
	}
	if got := read("home.toml"); got != "[[hero_images]]\nname = \"larch-1\"\nwidth = 2560\nheight = 1707\n" {
		t.Errorf("home config = %q", got)
	}
	if got := read("server.toml"); got != server {
		t.Errorf("server config was modified: %q", got)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown = %v, want none", unknown)
	}
	if got := fileNames(t, dir); !slices.Equal(got, []string{"home.toml", "land-gallery.toml", "server.toml"}) {
		t.Errorf("config directory contains %v, want no leftover temp files", got)
	}
	info, err := os.Stat(filepath.Join(dir, "land-gallery.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("gallery config mode = %v, want the original 0600", info.Mode().Perm())
	}
}

func TestUpdateConfigsReportsEachUnknownImageOnce(t *testing.T) {
	dir := t.TempDir()
	gallery := "[[rows.images]]\nname = \"missing\"\n"
	home := "[[hero_images]]\nname = \"missing\"\n"
	for name, content := range map[string]string{"a-gallery.toml": gallery, "home.toml": home} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	unknown, err := UpdateConfigs(dir, map[string]Dimensions{})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(unknown, []string{"missing"}) {
		t.Errorf("unknown = %v, want [missing]", unknown)
	}
}

func TestUpdateConfigsWorksWithoutAHomeConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a-gallery.toml"), []byte("[[rows.images]]\nname = \"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := UpdateConfigs(dir, map[string]Dimensions{"x": {Width: 10, Height: 5}}); err != nil {
		t.Errorf("UpdateConfigs without home.toml returned %v", err)
	}
}

func TestUpdateConfigsErrorsForMissingDirectory(t *testing.T) {
	if _, err := UpdateConfigs(filepath.Join(t.TempDir(), "missing"), nil); err == nil {
		t.Error("expected an error for a missing config directory")
	}
}
