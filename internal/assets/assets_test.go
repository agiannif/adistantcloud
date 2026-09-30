package assets

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, root, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func regexpForHashed(stem, ext string) *regexp.Regexp {
	return regexp.MustCompile(`^` + stem + `\.[0-9a-f]{8}\.` + ext + `$`)
}

func exists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

func TestFingerprintCopiesFileToHashedName(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "css/style.min.css", "body{}")

	manifest, err := Fingerprint(root, []string{"css/style.min.css"})
	if err != nil {
		t.Fatal(err)
	}

	hashed := manifest["css/style.min.css"]
	if hashed == "" || hashed == "css/style.min.css" {
		t.Fatalf("expected a hashed name in the manifest, got %q", hashed)
	}
	if got := readFile(t, root, hashed); got != "body{}" {
		t.Errorf("hashed copy content = %q, want %q", got, "body{}")
	}
	if !exists(root, "css/style.min.css") {
		t.Error("original file should be left in place")
	}
}

func TestFingerprintKeepsExtensionAndDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "script/alpinejs.min.js", "alpine")

	manifest, err := Fingerprint(root, []string{"script/alpinejs.min.js"})
	if err != nil {
		t.Fatal(err)
	}

	hashed := manifest["script/alpinejs.min.js"]
	want := regexpForHashed(`script/alpinejs\.min`, `js`)
	if !want.MatchString(hashed) {
		t.Errorf("hashed name %q does not match %s", hashed, want)
	}
}

func TestFingerprintNameIsStableForSameContent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "css/a.css", "same")

	first, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}

	if first["css/a.css"] != second["css/a.css"] {
		t.Errorf("name changed for identical content: %q vs %q", first["css/a.css"], second["css/a.css"])
	}
}

func TestFingerprintNameChangesWhenContentChanges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "css/a.css", "one")
	first, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "css/a.css", "two")
	second, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}

	if first["css/a.css"] == second["css/a.css"] {
		t.Errorf("name did not change when content changed: %q", first["css/a.css"])
	}
}

func TestFingerprintRemovesStaleHashedCopies(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "css/a.css", "one")
	first, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, root, "css/a.css", "two")
	second, err := Fingerprint(root, []string{"css/a.css"})
	if err != nil {
		t.Fatal(err)
	}

	if exists(root, first["css/a.css"]) {
		t.Errorf("stale hashed copy %q was not removed", first["css/a.css"])
	}
	if !exists(root, second["css/a.css"]) {
		t.Errorf("current hashed copy %q is missing", second["css/a.css"])
	}
}

func TestFingerprintLeavesUnrelatedFilesAlone(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "css/style.min.css", "body{}")
	writeFile(t, root, "css/style.min.other.css", "other")
	writeFile(t, root, "css/style.min.deadbeef.css.map", "map")
	writeFile(t, root, "css/style.min.notahash.css", "unrelated")

	if _, err := Fingerprint(root, []string{"css/style.min.css"}); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"css/style.min.css",
		"css/style.min.other.css",
		"css/style.min.deadbeef.css.map",
		"css/style.min.notahash.css",
	} {
		if !exists(root, name) {
			t.Errorf("%s should not have been removed", name)
		}
	}
}

func TestFingerprintErrorsOnMissingFile(t *testing.T) {
	root := t.TempDir()

	if _, err := Fingerprint(root, []string{"css/missing.css"}); err == nil {
		t.Fatal("expected an error for a missing source file")
	}
}

func TestManifestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	want := Manifest{"css/a.css": "css/a.deadbeef.css"}

	if err := want.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}

	if got["css/a.css"] != "css/a.deadbeef.css" {
		t.Errorf("round trip lost entry: %v", got)
	}
}

func TestLoadManifestErrorsWhenFileMissing(t *testing.T) {
	if _, err := LoadManifest(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected an error for a missing manifest")
	}
}

func TestManifestPathReturnsHashedNameWhenKnown(t *testing.T) {
	manifest := Manifest{"css/a.css": "css/a.deadbeef.css"}

	if got := manifest.Path("css/a.css"); got != "css/a.deadbeef.css" {
		t.Errorf("Path = %q, want hashed name", got)
	}
}

func TestManifestPathFallsBackToOriginalName(t *testing.T) {
	var manifest Manifest

	if got := manifest.Path("css/a.css"); got != "css/a.css" {
		t.Errorf("Path = %q, want the original name", got)
	}
}
