// Command imgprep turns exported PNG originals into the AVIF variants the site
// serves and records each original's dimensions in the gallery and home configs.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/agiannif/adistantcloud/internal/imageprep"
)

func main() {
	originals := flag.String("originals", "originals", "directory of PNG originals")
	output := flag.String("output", "assets/images", "directory the AVIF variants are written to")
	configs := flag.String("configs", "configs", "directory of gallery and home configs to record dimensions in")
	quality := flag.Int("quality", 70, "AVIF quality from 0 to 100, where 100 is lossless")
	speed := flag.Int("speed", 6, "avifenc encoder speed from 0, the slowest and best compressing, to 10")
	jobs := flag.Int("jobs", runtime.NumCPU(), "number of photos to process at once")
	force := flag.Bool("force", false, "regenerate variants even when they are newer than their original")
	flag.Parse()

	options := imageprep.Options{Quality: *quality, Speed: *speed, Jobs: *jobs, Force: *force}
	if err := run(*originals, *output, *configs, options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(originals, output, configs string, options imageprep.Options) error {
	paths, err := imageprep.OriginalPaths(originals)
	if err != nil {
		return err
	}

	// results are reported as photos finish, so the count is not the position in paths
	finished := 0
	results, err := imageprep.GenerateAll(paths, output, options, func(result imageprep.Result) {
		finished++
		status := ""
		if !result.Generated {
			status = " (up to date)"
		}
		fmt.Printf("[%d/%d] %s%s\n", finished, len(paths), result.Name, status)
	})
	if err != nil {
		return err
	}

	dims := make(map[string]imageprep.Dimensions, len(results))
	generated := 0
	for _, result := range results {
		dims[result.Name] = result.Dimensions
		if result.Generated {
			generated++
		}
	}
	fmt.Printf("generated %d, up to date %d\n", generated, len(results)-generated)

	unknown, err := imageprep.UpdateConfigs(configs, dims)
	if err != nil {
		return err
	}
	for _, name := range unknown {
		fmt.Fprintf(os.Stderr, "warning: %s is listed in the configs but has no original in %s\n", name, originals)
	}
	return nil
}
