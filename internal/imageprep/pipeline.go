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
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

// Options controls how originals are turned into AVIF variants.
type Options struct {
	// Quality is the AVIF quality from 0 to 100, where 100 is lossless.
	Quality int
	// Speed is the avifenc encoder speed from 0, the slowest and best compressing,
	// to 10.
	Speed int
	// Threads is how many threads each avifenc process may use. Zero lets avifenc
	// use every core, or lets GenerateAll divide the cores between its jobs.
	Threads int
	// Force regenerates variants that are already up to date.
	Force bool
	// Jobs is how many originals GenerateAll processes at once. Zero means one.
	Jobs int
}

// validate rejects values avifenc would silently clamp or misread.
func (o Options) validate() error {
	switch {
	case o.Quality < 0 || o.Quality > 100:
		return fmt.Errorf("quality must be between 0 and 100, got %d", o.Quality)
	case o.Speed < 0 || o.Speed > 10:
		return fmt.Errorf("speed must be between 0 and 10, got %d", o.Speed)
	case o.Threads < 0:
		return fmt.Errorf("threads must not be negative, got %d", o.Threads)
	case o.Jobs < 0:
		return fmt.Errorf("jobs must not be negative, got %d", o.Jobs)
	}
	return nil
}

// EncodeAVIF writes img to path as an AVIF using the avifenc command-line tool.
func EncodeAVIF(img image.Image, path string, options Options) error {
	if err := options.validate(); err != nil {
		return err
	}
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

	args := []string{"-q", strconv.Itoa(options.Quality), "-s", strconv.Itoa(options.Speed)}
	if options.Threads != 0 {
		args = append(args, "-j", strconv.Itoa(options.Threads))
	}
	output, err := exec.Command(avifenc, append(args, intermediate.Name(), path)...).CombinedOutput()
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
// unless options.Force is set.
func GenerateVariants(path, outputDir string, options Options) (Result, error) {
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
	if !options.Force && variantsUpToDate(info.ModTime(), outputDir, name, widths) {
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
		if err := EncodeAVIF(variant, filepath.Join(outputDir, images.Filename(name, width)), options); err != nil {
			return Result{}, err
		}
	}
	result.Generated = true
	return result, nil
}

// GenerateAll runs GenerateVariants for every path, processing options.Jobs
// originals at once, and returns the results in the order of paths. onResult, if
// not nil, is called once per original as it finishes and never from two
// goroutines at the same time. After the first failure no further originals are
// started and that failure is returned.
func GenerateAll(paths []string, outputDir string, options Options, onResult func(Result)) ([]Result, error) {
	jobs := min(max(options.Jobs, 1), max(len(paths), 1))
	if options.Threads == 0 && jobs > 1 {
		// share the cores between the concurrent encoders instead of oversubscribing them
		options.Threads = max(1, runtime.NumCPU()/jobs)
	}

	indexes := make(chan int, len(paths))
	for i := range paths {
		indexes <- i
	}
	close(indexes)

	var (
		results  = make([]Result, len(paths))
		mu       sync.Mutex
		firstErr error
		failed   atomic.Bool
		wg       sync.WaitGroup
	)
	for range jobs {
		wg.Go(func() {
			for i := range indexes {
				if failed.Load() {
					continue
				}
				result, err := GenerateVariants(paths[i], outputDir, options)

				mu.Lock()
				if err != nil {
					if firstErr == nil {
						firstErr = err
					}
					failed.Store(true)
				} else {
					results[i] = result
					if onResult != nil {
						onResult(result)
					}
				}
				mu.Unlock()
			}
		})
	}
	wg.Wait()

	return results, firstErr
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
