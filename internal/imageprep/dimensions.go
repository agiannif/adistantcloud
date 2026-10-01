// Package imageprep turns exported original photographs into the AVIF variants
// the site serves and records their dimensions in the site's TOML configs.
package imageprep

import (
	"fmt"
	"regexp"
	"strings"
)

// Dimensions are the pixel dimensions of an exported original.
type Dimensions struct {
	Width  int
	Height int
}

// imageTableHeaders are the TOML tables whose entries describe images.
var imageTableHeaders = []string{"[[rows.images]]", "[[hero_images]]"}

var (
	nameLine      = regexp.MustCompile(`^(\s*)name\s*=\s*"([^"]*)"`)
	dimensionLine = regexp.MustCompile(`^(\s*)(width|height)\s*=`)
)

// RecordDimensions sets width and height on every image table in a TOML config
// whose name has an entry in dims. It edits the text line by line rather than
// re-encoding the document so comments, ordering and spacing survive. It
// returns the names of images that have no entry in dims, which are left as is.
func RecordDimensions(content string, dims map[string]Dimensions) (updated string, unknown []string) {
	lines := strings.Split(content, "\n")
	var out []string

	for i := 0; i < len(lines); {
		out = append(out, lines[i])
		if !isImageTableHeader(lines[i]) {
			i++
			continue
		}

		end := i + 1
		for end < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[end]), "[") {
			end++
		}
		block, missing := recordBlockDimensions(lines[i+1:end], dims)
		out = append(out, block...)
		unknown = append(unknown, missing...)
		i = end
	}

	return strings.Join(out, "\n"), unknown
}

func isImageTableHeader(line string) bool {
	line = strings.TrimSpace(line)
	for _, header := range imageTableHeaders {
		if line == header {
			return true
		}
	}
	return false
}

// recordBlockDimensions updates the lines of one image table. Existing width and
// height lines are rewritten in place and missing ones are inserted after the
// name line.
func recordBlockDimensions(block []string, dims map[string]Dimensions) (updated []string, unknown []string) {
	nameIndex, name := -1, ""
	for i, line := range block {
		if match := nameLine.FindStringSubmatch(line); match != nil {
			nameIndex, name = i, match[2]
			break
		}
	}
	if nameIndex < 0 {
		return block, nil
	}

	dimension, ok := dims[name]
	if !ok {
		return block, []string{name}
	}

	values := map[string]int{"width": dimension.Width, "height": dimension.Height}
	existing := map[string]bool{}
	for _, line := range block {
		if match := dimensionLine.FindStringSubmatch(line); match != nil {
			existing[match[2]] = true
		}
	}

	indent := nameLine.FindStringSubmatch(block[nameIndex])[1]
	for i, line := range block {
		if match := dimensionLine.FindStringSubmatch(line); match != nil {
			line = fmt.Sprintf("%s%s = %d", match[1], match[2], values[match[2]])
		}
		updated = append(updated, line)
		if i != nameIndex {
			continue
		}
		for _, key := range []string{"width", "height"} {
			if !existing[key] {
				updated = append(updated, fmt.Sprintf("%s%s = %d", indent, key, values[key]))
			}
		}
	}
	return updated, nil
}
