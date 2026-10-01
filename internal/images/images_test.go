package images

import (
	"slices"
	"testing"
)

func TestWidths(t *testing.T) {
	tests := []struct {
		name        string
		sourceWidth int
		want        []int
	}{
		{"wide source gets every tier", 2560, []int{640, 1280, 2000}},
		{"source equal to the largest tier", 2000, []int{640, 1280, 2000}},
		{"narrower source caps the largest tier at its own width", 1707, []int{640, 1280, 1707}},
		{"source between two tiers", 1000, []int{640, 1000}},
		{"source equal to the smallest tier", 640, []int{640}},
		{"source narrower than every tier is never upscaled", 500, []int{500}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Widths(tt.sourceWidth); !slices.Equal(got, tt.want) {
				t.Errorf("Widths(%d) = %v, want %v", tt.sourceWidth, got, tt.want)
			}
		})
	}
}

func TestFilename(t *testing.T) {
	if got := Filename("larch-1", 1280); got != "larch-1-1280.avif" {
		t.Errorf("Filename = %q, want %q", got, "larch-1-1280.avif")
	}
}
