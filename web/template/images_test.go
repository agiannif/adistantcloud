package template

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/agiannif/adistantcloud/internal/config"
)

func TestImageSrcset(t *testing.T) {
	image := config.ImageConfig{Name: "larch-1", Width: 2560, Height: 1707}

	want := "/assets/images/larch-1-640.avif 640w, /assets/images/larch-1-1280.avif 1280w, /assets/images/larch-1-2000.avif 2000w"
	if got := imageSrcset(image); got != want {
		t.Errorf("imageSrcset = %q, want %q", got, want)
	}
}

func TestImageSrcsetForNarrowSourceUsesItsOwnWidth(t *testing.T) {
	image := config.ImageConfig{Name: "tall", Width: 1707, Height: 2560}

	want := "/assets/images/tall-640.avif 640w, /assets/images/tall-1280.avif 1280w, /assets/images/tall-1707.avif 1707w"
	if got := imageSrcset(image); got != want {
		t.Errorf("imageSrcset = %q, want %q", got, want)
	}
}

func TestLargestImageURL(t *testing.T) {
	image := config.ImageConfig{Name: "larch-1", Width: 2560, Height: 1707}

	if got := largestImageURL(image); got != "/assets/images/larch-1-2000.avif" {
		t.Errorf("largestImageURL = %q, want the 2000px variant", got)
	}
}

func TestImageSizes(t *testing.T) {
	tests := []struct {
		name string
		span string
		want string
	}{
		{"full width cell", "79", "(min-width: 48rem) calc(min(100vw, 80rem) * 1.00), 100vw"},
		{"half width cell", "39", "(min-width: 48rem) calc(min(100vw, 80rem) * 0.49), 100vw"},
		{"large split cell", "54", "(min-width: 48rem) calc(min(100vw, 80rem) * 0.68), 100vw"},
		{"small split cell", "24", "(min-width: 48rem) calc(min(100vw, 80rem) * 0.30), 100vw"},
		{"invalid span falls back to the viewport width", "wide", "100vw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imageSizes(tt.span); got != tt.want {
				t.Errorf("imageSizes(%q) = %q, want %q", tt.span, got, tt.want)
			}
		})
	}
}

func TestGalleryRendersResponsiveImages(t *testing.T) {
	gallery := &config.GalleryConfig{
		Metadata: config.GalleryMetadata{Name: "Test", ShortName: "test"},
		Rows: []config.RowConfig{
			// the first image is loaded eagerly, so the lazy markup is checked on later rows
			{Layout: config.LayoutFull, Images: []config.ImageConfig{{Name: "first", Width: 2560, Height: 1707}}},
			{
				Layout: config.LayoutSplit,
				Images: []config.ImageConfig{
					{Name: "wide", Alt: "a wide photo", Width: 2560, Height: 1707},
					{Name: "tall", Alt: "a tall photo", Width: 1707, Height: 2560},
				},
			},
		},
	}

	var page bytes.Buffer
	if err := Gallery(gallery, nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}
	html := page.String()

	for _, want := range []string{
		// lazy loading sets srcset, with sizes derived from the grid cell
		"$el.srcset = &#39;/assets/images/wide-640.avif 640w, /assets/images/wide-1280.avif 1280w, /assets/images/wide-2000.avif 2000w&#39;",
		`sizes="(min-width: 48rem) calc(min(100vw, 80rem) * 0.68), 100vw"`,
		`sizes="(min-width: 48rem) calc(min(100vw, 80rem) * 0.30), 100vw"`,
		// real dimensions reserve the right amount of space
		`width="2560" height="1707"`,
		`width="1707" height="2560"`,
		// the fullscreen viewer loads the largest variant
		"$el.src = &#39;/assets/images/wide-2000.avif&#39;",
		"$el.src = &#39;/assets/images/tall-1707.avif&#39;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered gallery does not contain %s", want)
		}
	}

	// Cells fill their grid track. An image that has not loaded yet has no src, so
	// it sizes itself from its width attribute; a cell that shrink-wraps it would
	// be as wide as the original, overflow the page, and make lazy loading think
	// the viewport is huge.
	if got := strings.Count(html, "bg-ef-light-bg-dim w-full"); got != 3 {
		t.Errorf("rendered gallery has %d image cells that fill their grid track, want 3", got)
	}

	// every lazily loaded thumbnail and every fullscreen image is decoded off the
	// main thread, the eagerly loaded first thumbnail is not
	if got := strings.Count(html, `decoding="async"`); got != 5 {
		t.Errorf(`rendered gallery has %d images with decoding="async", want 5`, got)
	}
}

func TestHomeRendersResponsiveHeroImage(t *testing.T) {
	hero := &config.ImageConfig{Name: "bay-1", Width: 2560, Height: 1707}

	var page bytes.Buffer
	if err := Home(hero, nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}
	html := page.String()

	for _, want := range []string{
		`srcset="/assets/images/bay-1-640.avif 640w, /assets/images/bay-1-1280.avif 1280w, /assets/images/bay-1-2000.avif 2000w"`,
		`sizes="(min-width: 80rem) 1232px, calc(100vw - 3rem)"`,
		`src="/assets/images/bay-1-2000.avif"`,
		// the hero is the largest thing on screen, so fetch it ahead of everything else
		`fetchpriority="high"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered home page does not contain %s", want)
		}
	}

	// the hero is painted as soon as it arrives rather than decoded asynchronously
	if strings.Contains(html, "decoding=") {
		t.Error("the hero image should not set a decoding hint")
	}
}

func TestHomeWithoutHeroImageRendersNoImage(t *testing.T) {
	var page bytes.Buffer
	if err := Home(nil, nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(page.String(), "/assets/images/") {
		t.Error("home page without a hero image should not reference any gallery image")
	}
}

func TestFirstImageRowIndex(t *testing.T) {
	images := []config.ImageConfig{{Name: "a", Width: 10, Height: 5}}
	tests := []struct {
		name string
		rows []config.RowConfig
		want int
	}{
		{"no rows", nil, -1},
		{"only section headers", []config.RowConfig{{Layout: config.LayoutSection}, {Layout: config.LayoutSection}}, -1},
		{"first row has images", []config.RowConfig{{Layout: config.LayoutFull, Images: images}}, 0},
		{"skips a leading section header", []config.RowConfig{{Layout: config.LayoutSection}, {Layout: config.LayoutHalf, Images: images}}, 1},
		{"skips a row without images", []config.RowConfig{{Layout: config.LayoutFull}, {Layout: config.LayoutFull, Images: images}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstImageRowIndex(tt.rows); got != tt.want {
				t.Errorf("firstImageRowIndex = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGalleryLoadsOnlyTheFirstImageEagerly(t *testing.T) {
	gallery := &config.GalleryConfig{
		Metadata: config.GalleryMetadata{Name: "Test", ShortName: "test"},
		Rows: []config.RowConfig{
			{Layout: config.LayoutSection, Title: "first section"},
			{Layout: config.LayoutHalf, Images: []config.ImageConfig{
				{Name: "first", Alt: "first photo", Width: 2560, Height: 1707},
				{Name: "second", Alt: "second photo", Width: 2560, Height: 1707},
			}},
			{Layout: config.LayoutFull, Images: []config.ImageConfig{{Name: "third", Width: 2560, Height: 1707}}},
		},
	}

	var page bytes.Buffer
	if err := Gallery(gallery, nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}
	html := page.String()

	// the first image is in the markup, so the browser can start fetching it
	// before any script runs, and it is fetched ahead of everything else
	tags := regexp.MustCompile(`<img [^>]*fetchpriority="high"[^>]*>`).FindAllString(html, -1)
	if len(tags) != 1 {
		t.Fatalf("found %d images with fetchpriority=high, want exactly the first", len(tags))
	}
	eager := tags[0]
	for _, want := range []string{
		`srcset="/assets/images/first-640.avif 640w, /assets/images/first-1280.avif 1280w, /assets/images/first-2000.avif 2000w"`,
		`sizes="(min-width: 48rem) calc(min(100vw, 80rem) * 0.49), 100vw"`,
		`alt="first photo"`,
		`width="2560" height="1707"`,
	} {
		if !strings.Contains(eager, want) {
			t.Errorf("eager image %s does not contain %s", eager, want)
		}
	}

	// it may finish loading before Alpine starts, so it must not wait for Alpine's
	// load handler to become visible, and it needs no lazy loading
	for _, unwanted := range []string{"opacity-0", "x-intersect", "decoding="} {
		if strings.Contains(eager, unwanted) {
			t.Errorf("eager image %s should not contain %s", eager, unwanted)
		}
	}

	// every other thumbnail stays lazily loaded
	for _, name := range []string{"second", "third"} {
		want := "$el.srcset = &#39;/assets/images/" + name + "-640.avif"
		if !strings.Contains(html, want) {
			t.Errorf("%s is not lazily loaded, rendered gallery does not contain %s", name, want)
		}
	}
	if strings.Contains(html, "$el.srcset = &#39;/assets/images/first-") {
		t.Error("the first image must not also be loaded lazily")
	}
}
