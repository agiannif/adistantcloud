package template

import (
	"testing"

	"github.com/agiannif/adistantcloud/internal/config"
)

func TestGetGridConfigs(t *testing.T) {
	tests := []struct {
		name   string
		layout config.Layout
		want   []GridConfig
	}{
		{
			name:   "LayoutFull returns full width grid",
			layout: config.LayoutFull,
			want:   []GridConfig{{"1", "79"}},
		},
		{
			name:   "LayoutHalf returns two half-width grids",
			layout: config.LayoutHalf,
			want:   []GridConfig{{"1", "39"}, {"41", "39"}},
		},
		{
			name:   "LayoutSplit returns large-small split",
			layout: config.LayoutSplit,
			want:   []GridConfig{{"1", "54"}, {"56", "24"}},
		},
		{
			name:   "LayoutFlipSplit returns small-large split",
			layout: config.LayoutFlipSplit,
			want:   []GridConfig{{"1", "24"}, {"26", "54"}},
		},
		{
			name:   "LayoutCenter returns centered grid",
			layout: config.LayoutCenter,
			want:   []GridConfig{{"21", "39"}},
		},
		{
			name:   "LayoutSection returns empty grid",
			layout: config.LayoutSection,
			want:   []GridConfig{},
		},
		{
			name:   "Unknown layout returns empty grid",
			layout: config.Layout("unknown"),
			want:   []GridConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGridConfigs(tt.layout)
			if len(got) != len(tt.want) {
				t.Errorf("getGridConfigs() returned %d configs, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Start != tt.want[i].Start || got[i].Span != tt.want[i].Span {
					t.Errorf("getGridConfigs()[%d] = {Start: %s, Span: %s}, want {Start: %s, Span: %s}",
						i, got[i].Start, got[i].Span, tt.want[i].Start, tt.want[i].Span)
				}
			}
		})
	}
}

func TestLayoutGridConfigs(t *testing.T) {
	t.Run("all defined layouts have grid configs", func(t *testing.T) {
		definedLayouts := []config.Layout{
			config.LayoutFull,
			config.LayoutHalf,
			config.LayoutSplit,
			config.LayoutFlipSplit,
			config.LayoutCenter,
		}

		for _, layout := range definedLayouts {
			if _, exists := layoutGridConfigs[layout]; !exists {
				t.Errorf("layout %q is missing from layoutGridConfigs map", layout)
			}
		}
	})

	t.Run("grid configs sum to 79 columns or less", func(t *testing.T) {
		for layout, configs := range layoutGridConfigs {
			if len(configs) == 0 {
				continue
			}

			for i, cfg := range configs {
				var start, span int
				if _, err := parseIntString(cfg.Start, &start); err != nil {
					t.Errorf("layout %q config[%d] has invalid Start: %q", layout, i, cfg.Start)
				}
				if _, err := parseIntString(cfg.Span, &span); err != nil {
					t.Errorf("layout %q config[%d] has invalid Span: %q", layout, i, cfg.Span)
				}

				end := start + span - 1
				if end > 79 {
					t.Errorf("layout %q config[%d] exceeds 79 columns: start=%d span=%d end=%d",
						layout, i, start, span, end)
				}
			}
		}
	})
}

func parseIntString(s string, out *int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	*out = n
	return n, nil
}
