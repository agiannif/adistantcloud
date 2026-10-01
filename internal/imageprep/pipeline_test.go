package imageprep

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
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

	if err := EncodeAVIF(gradient(64, 48), path, testOptions); err != nil {
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

	if err := EncodeAVIF(img, filepath.Join(dir, "high.avif"), Options{Quality: 90, Speed: 6}); err != nil {
		t.Fatal(err)
	}
	if err := EncodeAVIF(img, filepath.Join(dir, "low.avif"), Options{Quality: 10, Speed: 6}); err != nil {
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

	if err := EncodeAVIF(gradient(64, 48), path, testOptions); err == nil {
		t.Error("expected an error when the output directory does not exist")
	}
}

func TestEncodeAVIFErrorsClearlyWhenAvifencIsNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := EncodeAVIF(gradient(64, 48), filepath.Join(t.TempDir(), "out.avif"), testOptions)

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

	result, err := GenerateVariants(path, out, testOptions)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != "wide" {
		t.Errorf("name = %q, want %q", result.Name, "wide")
	}
	if result.Dimensions != (Dimensions{Width: 800, Height: 400}) {
		t.Errorf("dimensions = %+v, want 800x400", result.Dimensions)
	}
	if !result.Generated {
		t.Error("Generated = false, want true for a new original")
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

	if _, err := GenerateVariants(path, out, testOptions); err != nil {
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

	_, err := GenerateVariants(path, t.TempDir(), testOptions)

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

// generatedFixture generates variants for a fresh 800x400 original and returns
// the original's path, the output directory and the path of one variant.
func generatedFixture(t *testing.T) (original, out, variant string) {
	t.Helper()
	requireAvifenc(t)
	original, out = filepath.Join(t.TempDir(), "wide.png"), t.TempDir()
	writePNG(t, original, gradient(800, 400))
	if _, err := GenerateVariants(original, out, testOptions); err != nil {
		t.Fatal(err)
	}
	return original, out, filepath.Join(out, "wide-640.avif")
}

func setModTime(t *testing.T, path string, offset time.Duration) {
	t.Helper()
	when := time.Now().Add(offset)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
}

func replaceWithMarker(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("marker"), 0o644); err != nil {
		t.Fatal(err)
	}
	setModTime(t, path, 0)
}

func isMarker(t *testing.T, path string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data) == "marker"
}

func TestGenerateVariantsSkipsWhenEveryVariantIsNewerThanTheOriginal(t *testing.T) {
	original, out, variant := generatedFixture(t)
	setModTime(t, original, -time.Hour)
	replaceWithMarker(t, variant)

	result, err := GenerateVariants(original, out, testOptions)
	if err != nil {
		t.Fatal(err)
	}

	if result.Generated {
		t.Error("Generated = true, want false when the variants are up to date")
	}
	if !isMarker(t, variant) {
		t.Error("an up to date variant was rewritten")
	}
	if result.Name != "wide" || result.Dimensions != (Dimensions{Width: 800, Height: 400}) {
		t.Errorf("a skipped original must still report its name and dimensions, got %+v", result)
	}
}

func TestGenerateVariantsRegeneratesWhenTheOriginalIsNewer(t *testing.T) {
	original, out, variant := generatedFixture(t)
	replaceWithMarker(t, variant)
	setModTime(t, variant, -time.Hour)
	setModTime(t, original, 0)

	result, err := GenerateVariants(original, out, testOptions)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Generated {
		t.Error("Generated = false, want true when the original is newer")
	}
	if isMarker(t, variant) {
		t.Error("a stale variant was not rewritten")
	}
}

func TestGenerateVariantsRegeneratesWhenAVariantIsMissing(t *testing.T) {
	original, out, variant := generatedFixture(t)
	setModTime(t, original, -time.Hour)
	if err := os.Remove(variant); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateVariants(original, out, testOptions)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Generated {
		t.Error("Generated = false, want true when a variant is missing")
	}
	if _, err := os.Stat(variant); err != nil {
		t.Errorf("missing variant was not regenerated: %v", err)
	}
}

func TestGenerateVariantsForceRegeneratesUpToDateVariants(t *testing.T) {
	original, out, variant := generatedFixture(t)
	setModTime(t, original, -time.Hour)
	replaceWithMarker(t, variant)

	result, err := GenerateVariants(original, out, Options{Quality: 50, Speed: 6, Force: true})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Generated {
		t.Error("Generated = false, want true when forced")
	}
	if isMarker(t, variant) {
		t.Error("a forced run did not rewrite the variant")
	}
}

func TestGenerateVariantsChecksTheColorSpaceEvenWhenSkipping(t *testing.T) {
	original, out, _ := generatedFixture(t)
	if err := os.WriteFile(original, pngWithChunks(t, iccProfile("Display P3")), 0o644); err != nil {
		t.Fatal(err)
	}
	setModTime(t, original, -time.Hour)

	if _, err := GenerateVariants(original, out, testOptions); err == nil {
		t.Error("expected a Display P3 original to be refused even though its variants are up to date")
	}
}

func TestOriginalPathsSortsNumbersNaturally(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"tokyo-10.png", "tokyo-2.png", "tokyo-1.png", "china-1.png", "tokyo-21.png"} {
		writePNG(t, filepath.Join(dir, name), gradient(8, 8))
	}

	got, err := OriginalPaths(dir)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, path := range got {
		names = append(names, filepath.Base(path))
	}
	want := []string{"china-1.png", "tokyo-1.png", "tokyo-2.png", "tokyo-10.png", "tokyo-21.png"}
	if !slices.Equal(names, want) {
		t.Errorf("OriginalPaths order = %v, want %v", names, want)
	}
}

func TestUpdateConfigsReportsUnknownImagesInNaturalOrder(t *testing.T) {
	dir := t.TempDir()
	gallery := "[[rows.images]]\nname = \"x-10\"\n[[rows.images]]\nname = \"x-2\"\n"
	if err := os.WriteFile(filepath.Join(dir, "a-gallery.toml"), []byte(gallery), 0o644); err != nil {
		t.Fatal(err)
	}

	unknown, err := UpdateConfigs(dir, map[string]Dimensions{})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(unknown, []string{"x-2", "x-10"}) {
		t.Errorf("unknown = %v, want [x-2 x-10]", unknown)
	}
}

var testOptions = Options{Quality: 50, Speed: 6}

func TestEncodeAVIFRejectsOptionsOutOfRange(t *testing.T) {
	// avifenc silently accepts most out-of-range values, so they are checked here
	tests := []struct {
		name    string
		options Options
		want    string
	}{
		{"quality above 100", Options{Quality: 101, Speed: 6}, "quality"},
		{"negative quality", Options{Quality: -1, Speed: 6}, "quality"},
		{"speed above 10", Options{Quality: 50, Speed: 11}, "speed"},
		{"negative speed", Options{Quality: 50, Speed: -1}, "speed"},
		{"negative threads", Options{Quality: 50, Speed: 6, Threads: -3}, "threads"},
		{"negative jobs", Options{Quality: 50, Speed: 6, Jobs: -2}, "jobs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out.avif")

			err := EncodeAVIF(gradient(64, 48), path, tt.options)

			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("EncodeAVIF = %v, want an error mentioning %q", err, tt.want)
			}
			if _, statErr := os.Stat(path); statErr == nil {
				t.Error("an AVIF was written despite the invalid options")
			}
		})
	}
}

func TestEncodeAVIFAcceptsTheLimitsOfEachOption(t *testing.T) {
	requireAvifenc(t)
	for _, options := range []Options{
		{Quality: 0, Speed: 10, Threads: 1},
		{Quality: 100, Speed: 0, Threads: 2},
	} {
		path := filepath.Join(t.TempDir(), "out.avif")
		if err := EncodeAVIF(gradient(32, 32), path, options); err != nil {
			t.Errorf("EncodeAVIF(%+v) returned %v", options, err)
		}
	}
}

func TestEncodeAVIFAppliesSpeed(t *testing.T) {
	requireAvifenc(t)
	dir := t.TempDir()
	img := gradient(512, 512)

	if err := EncodeAVIF(img, filepath.Join(dir, "slow.avif"), Options{Quality: 50, Speed: 0}); err != nil {
		t.Fatal(err)
	}
	if err := EncodeAVIF(img, filepath.Join(dir, "fast.avif"), Options{Quality: 50, Speed: 10}); err != nil {
		t.Fatal(err)
	}

	slow, _ := os.ReadFile(filepath.Join(dir, "slow.avif"))
	fast, _ := os.ReadFile(filepath.Join(dir, "fast.avif"))
	if bytes.Equal(slow, fast) {
		t.Error("speed 0 and speed 10 produced identical files, so the speed option is not being applied")
	}
}

// writeOriginals writes small PNG originals into a new directory and returns
// their paths in order.
func writeOriginals(t *testing.T, sizes map[string][2]int, order []string) (string, []string) {
	t.Helper()
	dir := t.TempDir()
	var paths []string
	for _, name := range order {
		path := filepath.Join(dir, name+".png")
		writePNG(t, path, gradient(sizes[name][0], sizes[name][1]))
		paths = append(paths, path)
	}
	return dir, paths
}

func TestGenerateAllGeneratesEveryOriginalAndKeepsInputOrder(t *testing.T) {
	requireAvifenc(t)
	sizes := map[string][2]int{"a": {800, 400}, "b": {300, 600}, "c": {700, 700}, "d": {900, 300}, "e": {650, 650}}
	order := []string{"a", "b", "c", "d", "e"}
	_, paths := writeOriginals(t, sizes, order)
	out := t.TempDir()

	results, err := GenerateAll(paths, out, Options{Quality: 50, Speed: 6, Jobs: 3}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != len(order) {
		t.Fatalf("got %d results, want %d", len(results), len(order))
	}
	for i, name := range order {
		if results[i].Name != name {
			t.Errorf("results[%d].Name = %q, want %q, results must follow the input order", i, results[i].Name, name)
		}
		want := Dimensions{Width: sizes[name][0], Height: sizes[name][1]}
		if results[i].Dimensions != want || !results[i].Generated {
			t.Errorf("results[%d] = %+v, want dimensions %+v and Generated", i, results[i], want)
		}
	}
	if got := len(fileNames(t, out)); got != 9 {
		t.Errorf("generated %d files, want 9 (a: 640 and 800, b: 300, c: 640 and 700, d: 640 and 900, e: 640 and 650)", got)
	}
}

func TestGenerateAllReportsEachResultOnceWithoutOverlapping(t *testing.T) {
	requireAvifenc(t)
	order := []string{"a", "b", "c", "d", "e", "f"}
	sizes := map[string][2]int{}
	for _, name := range order {
		sizes[name] = [2]int{200, 100}
	}
	_, paths := writeOriginals(t, sizes, order)

	// the callback deliberately uses plain, unsynchronised state: under -race any
	// overlapping calls would be reported
	reported := map[string]int{}
	var calls int
	_, err := GenerateAll(paths, t.TempDir(), Options{Quality: 50, Speed: 6, Jobs: 4}, func(result Result) {
		reported[result.Name]++
		calls++
	})
	if err != nil {
		t.Fatal(err)
	}

	if calls != len(order) {
		t.Errorf("callback ran %d times, want %d", calls, len(order))
	}
	for _, name := range order {
		if reported[name] != 1 {
			t.Errorf("%s reported %d times, want once", name, reported[name])
		}
	}
}

func TestGenerateAllTreatsZeroJobsAsOne(t *testing.T) {
	requireAvifenc(t)
	_, paths := writeOriginals(t, map[string][2]int{"a": {200, 100}}, []string{"a"})

	results, err := GenerateAll(paths, t.TempDir(), Options{Quality: 50, Speed: 6, Jobs: 0}, nil)

	if err != nil || len(results) != 1 || !results[0].Generated {
		t.Errorf("GenerateAll with Jobs 0 = %+v, %v; want one generated result", results, err)
	}
}

func TestGenerateAllStopsAfterAFailureAndNamesTheFile(t *testing.T) {
	requireAvifenc(t)
	dir, paths := writeOriginals(t, map[string][2]int{"a": {200, 100}, "c": {200, 100}, "d": {200, 100}}, []string{"a", "c", "d"})
	broken := filepath.Join(dir, "b.png")
	if err := os.WriteFile(broken, []byte("not a png"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths = []string{paths[0], broken, paths[1], paths[2]}
	out := t.TempDir()

	// one job makes the order deterministic: a, then the broken b, then nothing
	_, err := GenerateAll(paths, out, Options{Quality: 50, Speed: 6, Jobs: 1}, nil)

	if err == nil || !strings.Contains(err.Error(), "b.png") {
		t.Fatalf("error = %v, want one naming b.png", err)
	}
	for _, name := range fileNames(t, out) {
		if strings.HasPrefix(name, "c-") || strings.HasPrefix(name, "d-") {
			t.Errorf("%s was generated after the failure, remaining work should be skipped", name)
		}
	}
}

func TestGenerateAllSkipsUpToDateOriginals(t *testing.T) {
	requireAvifenc(t)
	_, paths := writeOriginals(t, map[string][2]int{"a": {200, 100}, "b": {200, 100}}, []string{"a", "b"})
	out := t.TempDir()
	options := Options{Quality: 50, Speed: 6, Jobs: 2}
	if _, err := GenerateAll(paths, out, options, nil); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		setModTime(t, path, -time.Hour)
	}

	results, err := GenerateAll(paths, out, options, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, result := range results {
		if result.Generated {
			t.Errorf("%s was regenerated although its variants were up to date", result.Name)
		}
	}
}
