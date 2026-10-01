package imageprep

import (
	"bytes"
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
	"time"

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

// OriginalPaths returns the PNG originals in dir, in natural order so img-2 comes
// before img-10.
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
	slices.SortFunc(paths, naturalCompare)
	return paths, nil
}

// Result describes what GenerateVariants did for one original.
type Result struct {
	// Name is the image's name, taken from the original's file name.
	Name string
	// Dimensions are the dimensions of the original.
	Dimensions Dimensions
	// Generated is false when the variants were already up to date.
	Generated bool
}

// GenerateVariants writes every AVIF variant of the original at path into
// outputDir. Variants that exist and are newer than the original are left alone
// unless force is set.
func GenerateVariants(path, outputDir string, quality int, force bool) (Result, error) {
	data, err := readOriginal(path)
	if err != nil {
		return Result{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	header, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Result{}, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	result := Result{Name: name, Dimensions: Dimensions{Width: header.Width, Height: header.Height}}
	widths := images.Widths(header.Width)
	if !force && variantsUpToDate(info.ModTime(), outputDir, name, widths) {
		return result, nil
	}

	original, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return Result{}, fmt.Errorf("failed to decode %s: %w", path, err)
	}
	for _, width := range widths {
		variant := original
		if width != header.Width {
			variant = Resize(original, width)
		}
		if err := EncodeAVIF(variant, filepath.Join(outputDir, images.Filename(name, width)), quality); err != nil {
			return Result{}, err
		}
	}
	result.Generated = true
	return result, nil
}

// variantsUpToDate reports whether every variant exists and is at least as new as
// the original.
func variantsUpToDate(originalModTime time.Time, outputDir, name string, widths []int) bool {
	for _, width := range widths {
		info, err := os.Stat(filepath.Join(outputDir, images.Filename(name, width)))
		if err != nil || info.ModTime().Before(originalModTime) {
			return false
		}
	}
	return true
}

// readOriginal reads a PNG original and refuses it unless it is sRGB.
func readOriginal(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	if err := checkSRGB(data); err != nil {
		return nil, fmt.Errorf("%s is not sRGB (%w), export it with the sRGB color space", path, err)
	}
	return data, nil
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

	slices.SortFunc(unknown, naturalCompare)
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
