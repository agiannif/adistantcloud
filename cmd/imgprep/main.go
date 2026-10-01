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
	force := flag.Bool("force", false, "regenerate variants even when they are newer than their original")
	flag.Parse()

	if err := run(*originals, *output, *configs, *quality, *force); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(originals, output, configs string, quality int, force bool) error {
	paths, err := imageprep.OriginalPaths(originals)
	if err != nil {
		return err
	}

	dims := make(map[string]imageprep.Dimensions, len(paths))
	generated := 0
	for i, path := range paths {
		result, err := imageprep.GenerateVariants(path, output, quality, force)
		if err != nil {
			return err
		}
		dims[result.Name] = result.Dimensions
		if result.Generated {
			generated++
			fmt.Printf("[%d/%d] %s\n", i+1, len(paths), result.Name)
		} else {
			fmt.Printf("[%d/%d] %s (up to date)\n", i+1, len(paths), result.Name)
		}
	}
	fmt.Printf("generated %d, up to date %d\n", generated, len(paths)-generated)

	unknown, err := imageprep.UpdateConfigs(configs, dims)
	if err != nil {
		return err
	}
	for _, name := range unknown {
		fmt.Fprintf(os.Stderr, "warning: %s is listed in the configs but has no original in %s\n", name, originals)
	}
	return nil
}
