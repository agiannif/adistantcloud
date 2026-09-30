// Package assets gives static files content-addressed names so they can be
// cached indefinitely by browsers.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const hashLength = 8

// Manifest maps a static file's name to its content-addressed name. Both are
// slash-separated and relative to the static root.
type Manifest map[string]string

// Path returns the content-addressed name for name, or name unchanged when the
// manifest does not know it.
func (m Manifest) Path(name string) string {
	if hashed, ok := m[name]; ok {
		return hashed
	}
	return name
}

func (m Manifest) Save(location string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode asset manifest: %w", err)
	}
	if err := os.WriteFile(location, data, 0o644); err != nil {
		return fmt.Errorf("failed to write asset manifest: %w", err)
	}
	return nil
}

func LoadManifest(location string) (Manifest, error) {
	data, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse asset manifest: %w", err)
	}
	return manifest, nil
}

// Fingerprint copies each file under root to a name containing a hash of its
// contents (style.css becomes style.3f9a1c2e.css) and removes hashed copies
// left over from earlier contents. The originals are kept.
func Fingerprint(root string, files []string) (Manifest, error) {
	manifest := Manifest{}
	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		hashed := hashedName(name, content)
		if err := removeStaleCopies(root, name); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(hashed)), content, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", hashed, err)
		}
		manifest[name] = hashed
	}
	return manifest, nil
}

func hashedName(name string, content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])[:hashLength]
	ext := path.Ext(name)
	return name[:len(name)-len(ext)] + "." + hash + ext
}

func removeStaleCopies(root, name string) error {
	ext := path.Ext(name)
	stem := path.Base(name[:len(name)-len(ext)])
	stale := regexp.MustCompile(`^` + regexp.QuoteMeta(stem) + `\.[0-9a-f]{` + fmt.Sprint(hashLength) + `}` + regexp.QuoteMeta(ext) + `$`)

	dir := filepath.Join(root, filepath.FromSlash(path.Dir(name)))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to list %s: %w", dir, err)
	}
	for _, entry := range entries {
		if !stale.MatchString(entry.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return fmt.Errorf("failed to remove stale copy %s: %w", entry.Name(), err)
		}
	}
	return nil
}
