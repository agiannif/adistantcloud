package template

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/agiannif/adistantcloud/internal/assets"
)

func useAssetManifest(t *testing.T, manifest assets.Manifest) {
	t.Helper()
	SetAssetManifest(manifest)
	t.Cleanup(func() { SetAssetManifest(nil) })
}

func TestAssetURLUsesHashedNameFromManifest(t *testing.T) {
	useAssetManifest(t, assets.Manifest{"css/style.min.css": "css/style.min.deadbeef.css"})

	if got := AssetURL("css/style.min.css"); got != "/static/css/style.min.deadbeef.css" {
		t.Errorf("AssetURL = %q, want hashed URL", got)
	}
}

func TestAssetURLFallsBackToOriginalNameWithoutManifest(t *testing.T) {
	useAssetManifest(t, nil)

	if got := AssetURL("css/style.min.css"); got != "/static/css/style.min.css" {
		t.Errorf("AssetURL = %q, want unversioned URL", got)
	}
}

func TestPageReferencesHashedAssets(t *testing.T) {
	useAssetManifest(t, assets.Manifest{
		"css/style.min.css":                "css/style.min.aaaaaaaa.css",
		"css/fonts.css":                    "css/fonts.bbbbbbbb.css",
		"script/alpinejs.min.js":           "script/alpinejs.min.cccccccc.js",
		"script/alpinejs-intersect.min.js": "script/alpinejs-intersect.min.dddddddd.js",
		"images/large_clouds.png":          "images/large_clouds.eeeeeeee.png",
	})

	var page bytes.Buffer
	if err := Home("", nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"/static/css/style.min.aaaaaaaa.css",
		"/static/css/fonts.bbbbbbbb.css",
		"/static/script/alpinejs.min.cccccccc.js",
		"/static/script/alpinejs-intersect.min.dddddddd.js",
		"/static/images/large_clouds.eeeeeeee.png",
	} {
		if !strings.Contains(page.String(), want) {
			t.Errorf("rendered page does not reference %s", want)
		}
	}
}
