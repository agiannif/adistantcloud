// Package images defines the sizes and file names of the AVIF variants generated
// for each photograph. It is shared by the web server and the image pipeline.
package images

import (
	"fmt"
	"slices"
)

// tiers are the widths in pixels that variants are generated at: small and half
// width cells, desktop full rows and phones, and retina full rows and fullscreen.
var tiers = []int{640, 1280, 2000}

// Widths returns the variant widths for a source image of the given width. A
// variant is never wider than its source, so a narrow source's largest variant
// is the source width itself.
func Widths(sourceWidth int) []int {
	widths := make([]int, 0, len(tiers))
	for _, tier := range tiers {
		width := min(tier, sourceWidth)
		if !slices.Contains(widths, width) {
			widths = append(widths, width)
		}
	}
	return widths
}

// Filename returns the file name of the variant of an image at the given width.
func Filename(name string, width int) string {
	return fmt.Sprintf("%s-%d.avif", name, width)
}
