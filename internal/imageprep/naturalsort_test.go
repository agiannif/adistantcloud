package imageprep

import (
	"slices"
	"testing"
)

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"smaller number first", "img-2", "img-10", -1},
		{"larger number last", "img-10", "img-2", 1},
		{"single digit before double digit", "img-9", "img-10", -1},
		{"identical strings", "img-3", "img-3", 0},
		{"leading zeros compare by value", "img-02", "img-10", -1},
		{"numerically equal strings still have a stable order", "img-01", "img-1", -1},
		{"text decides before numbers", "china-10", "tokyo-1", -1},
		{"prefix sorts first", "img", "img-1", -1},
		{"number in the middle", "a-2-b", "a-10-a", -1},
		{"very long numbers do not overflow", "n-99999999999999999999", "n-100000000000000000000", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := naturalCompare(tt.a, tt.b); got != tt.want {
				t.Errorf("naturalCompare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestNaturalCompareSortsAList(t *testing.T) {
	names := []string{"x-10", "x-1", "x-2", "x-21", "x-3"}

	slices.SortFunc(names, naturalCompare)

	want := []string{"x-1", "x-2", "x-3", "x-10", "x-21"}
	if !slices.Equal(names, want) {
		t.Errorf("sorted = %v, want %v", names, want)
	}
}
