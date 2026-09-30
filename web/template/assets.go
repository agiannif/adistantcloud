package template

import "github.com/agiannif/adistantcloud/internal/assets"

var assetManifest assets.Manifest

// SetAssetManifest sets the content-hashed names that AssetURL resolves to.
func SetAssetManifest(manifest assets.Manifest) {
	assetManifest = manifest
}

// AssetURL returns the URL of a static file, using its content-hashed name when
// the manifest has one so browsers can cache it indefinitely.
func AssetURL(name string) string {
	return "/static/" + assetManifest.Path(name)
}
