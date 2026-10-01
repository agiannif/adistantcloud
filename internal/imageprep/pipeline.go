package imageprep

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/image/draw"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/internal/images"
)

// Resize scales src to the given width, keeping its aspect ratio.
func Resize(src image.Image, width int) image.Image {
	bounds := src.Bounds()
	height := max(1, int(math.Round(float64(bounds.Dy())*float64(width)/float64(bounds.Dx()))))
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), src, bounds, draw.Over, nil)
	return resized
}

// EncodeAVIF writes img to path as an AVIF using the avifenc command-line tool.
// Quality is 0 to 100, where 100 is lossless.
func EncodeAVIF(img image.Image, path string, quality int) error {
	avifenc, err := exec.LookPath("avifenc")
	if err != nil {
		return errors.New("avifenc not found on PATH, install libavif (brew install libavif, apt install libavif-bin)")
	}

	intermediate, err := os.CreateTemp("", "imageprep-*.png")
	if err != nil {
		return fmt.Errorf("failed to create temporary PNG: %w", err)
	}
	defer os.Remove(intermediate.Name())

	// avifenc reads PNG, so hand it a fast, lossless intermediate
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(intermediate, img); err != nil {
		intermediate.Close()
		return fmt.Errorf("failed to write temporary PNG: %w", err)
	}
	if err := intermediate.Close(); err != nil {
		return fmt.Errorf("failed to write temporary PNG: %w", err)
	}

	output, err := exec.Command(avifenc, "-q", fmt.Sprint(quality), intermediate.Name(), path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("avifenc failed for %s: %w\n%s", path, err, output)
	}
	return nil
}

// OriginalPaths returns the PNG originals in dir, sorted by file name.
func OriginalPaths(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read originals directory %s: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no PNG originals found in %s", dir)
	}
	return paths, nil
}

// GenerateVariants writes every AVIF variant of the original at path into
// outputDir. It returns the image's name, taken from the file name, and the
// dimensions of the original.
func GenerateVariants(path, outputDir string, quality int) (string, Dimensions, error) {
	original, err := decodePNG(path)
	if err != nil {
		return "", Dimensions{}, err
	}

	bounds := original.Bounds()
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	for _, width := range images.Widths(bounds.Dx()) {
		variant := original
		if width != bounds.Dx() {
			variant = Resize(original, width)
		}
		if err := EncodeAVIF(variant, filepath.Join(outputDir, images.Filename(name, width)), quality); err != nil {
			return "", Dimensions{}, err
		}
	}
	return name, Dimensions{Width: bounds.Dx(), Height: bounds.Dy()}, nil
}

func decodePNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", path, err)
	}
	return img, nil
}

// UpdateConfigs records dimensions in the gallery configs and the home config in
// configsDir. It returns the sorted names of images the configs list that have no
// entry in dims.
func UpdateConfigs(configsDir string, dims map[string]Dimensions) ([]string, error) {
	files, err := config.GalleryConfigFiles(configsDir)
	if err != nil {
		return nil, err
	}
	home := filepath.Join(configsDir, "home.toml")
	if _, err := os.Stat(home); err == nil {
		files = append(files, home)
	}

	var unknown []string
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}

		updated, missing := RecordDimensions(string(content), dims)
		unknown = append(unknown, missing...)
		if updated == string(content) {
			continue
		}
		if err := writeFileAtomic(file, []byte(updated)); err != nil {
			return nil, err
		}
	}

	slices.Sort(unknown)
	return slices.Compact(unknown), nil
}

// writeFileAtomic replaces path with data so a failure cannot leave a partly
// written file, keeping the original file's permissions.
func writeFileAtomic(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".imageprep-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for %s: %w", path, err)
	}
	defer os.Remove(temp.Name())

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	if err := os.Chmod(temp.Name(), info.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("failed to replace %s: %w", path, err)
	}
	return nil
}
