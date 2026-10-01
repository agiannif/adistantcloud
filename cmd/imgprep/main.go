// Command imgprep turns exported PNG originals into the AVIF variants the site
// serves and records each original's dimensions in the gallery and home configs.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/agiannif/adistantcloud/internal/imageprep"
)

func main() {
	originals := flag.String("originals", "originals", "directory of PNG originals")
	output := flag.String("output", "assets/images", "directory the AVIF variants are written to")
	configs := flag.String("configs", "configs", "directory of gallery and home configs to record dimensions in")
	quality := flag.Int("quality", 50, "AVIF quality from 0 to 100, where 100 is lossless")
	flag.Parse()

	if err := run(*originals, *output, *configs, *quality); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(originals, output, configs string, quality int) error {
	paths, err := imageprep.OriginalPaths(originals)
	if err != nil {
		return err
	}

	dims := make(map[string]imageprep.Dimensions, len(paths))
	for i, path := range paths {
		name, dimensions, err := imageprep.GenerateVariants(path, output, quality)
		if err != nil {
			return err
		}
		dims[name] = dimensions
		fmt.Printf("[%d/%d] %s\n", i+1, len(paths), name)
	}

	unknown, err := imageprep.UpdateConfigs(configs, dims)
	if err != nil {
		return err
	}
	for _, name := range unknown {
		fmt.Fprintf(os.Stderr, "warning: %s is listed in the configs but has no original in %s\n", name, originals)
	}
	return nil
}
