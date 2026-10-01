package template

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/internal/images"
)

const (
	imageURLPrefix = "/assets/images/"

	// gridColumns is the number of columns in the gallery grid.
	gridColumns = 79

	// heroImageSizes matches the home page hero container: max-w-7xl with px-6.
	heroImageSizes = "(min-width: 80rem) 1232px, calc(100vw - 3rem)"
)

// imageSrcset lists every AVIF variant of an image with its width.
func imageSrcset(image config.ImageConfig) string {
	widths := images.Widths(image.Width)
	candidates := make([]string, len(widths))
	for i, width := range widths {
		candidates[i] = imageURLPrefix + images.Filename(image.Name, width) + " " + strconv.Itoa(width) + "w"
	}
	return strings.Join(candidates, ", ")
}

// largestImageURL returns the URL of an image's largest variant.
func largestImageURL(image config.ImageConfig) string {
	widths := images.Widths(image.Width)
	return imageURLPrefix + images.Filename(image.Name, widths[len(widths)-1])
}

// imageSizes tells the browser how wide a gallery image renders so it can pick
// a variant: a share of the content area (at most 80rem) from the md breakpoint
// up, the full viewport below it.
func imageSizes(span string) string {
	columns, err := strconv.Atoi(span)
	if err != nil {
		return "100vw"
	}
	share := float64(columns) / gridColumns
	return fmt.Sprintf("(min-width: 48rem) calc(min(100vw, 80rem) * %.2f), 100vw", share)
}

// firstImageRowIndex returns the index of the first row that has images, or -1
// when no row does.
func firstImageRowIndex(rows []config.RowConfig) int {
	return slices.IndexFunc(rows, func(row config.RowConfig) bool { return len(row.Images) > 0 })
}
