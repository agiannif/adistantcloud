package main

import (
	"fmt"
	"os"

	"github.com/agiannif/adistantcloud/internal/assets"
)

const (
	staticRoot   = "./web/static"
	manifestPath = staticRoot + "/assets-manifest.json"
)

// versioned lists the static files that templates reference through
// template.AssetURL. Fonts, favicons and the web manifest are not versioned.
var versioned = []string{
	"css/style.min.css",
	"css/fonts.css",
	"script/alpinejs.min.js",
	"script/alpinejs-intersect.min.js",
	"images/large_clouds.png",
}

func main() {
	manifest, err := assets.Fingerprint(staticRoot, versioned)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := manifest.Save(manifestPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
