package imageprep

import (
	"slices"
	"testing"
)

var testDimensions = map[string]Dimensions{
	"larch-1": {Width: 2560, Height: 1707},
	"tall":    {Width: 1707, Height: 2560},
}

func TestRecordDimensionsInsertsAfterName(t *testing.T) {
	content := `[[rows]]
layout = "full"
[[rows.images]]
name = "larch-1"
alt = "a larch"
`
	want := `[[rows]]
layout = "full"
[[rows.images]]
name = "larch-1"
width = 2560
height = 1707
alt = "a larch"
`

	got, unknown := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown = %v, want none", unknown)
	}
}

func TestRecordDimensionsUpdatesExistingValues(t *testing.T) {
	content := `[[rows.images]]
name = "larch-1"
width = 100
height = 50
`
	want := `[[rows.images]]
name = "larch-1"
width = 2560
height = 1707
`

	got, _ := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
}

func TestRecordDimensionsInsertsOnlyTheMissingKey(t *testing.T) {
	content := `[[rows.images]]
name = "larch-1"
width = 100
`
	want := `[[rows.images]]
name = "larch-1"
height = 1707
width = 2560
`

	got, _ := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
}

func TestRecordDimensionsIsIdempotent(t *testing.T) {
	content := `[[rows.images]]
name = "larch-1"
alt = "a larch"
`

	once, _ := RecordDimensions(content, testDimensions)
	twice, _ := RecordDimensions(once, testDimensions)

	if once != twice {
		t.Errorf("second run changed the content:\n%s\nvs\n%s", once, twice)
	}
}

func TestRecordDimensionsPreservesCommentsAndBlankLines(t *testing.T) {
	content := `[metadata]
name = "landscape"
short_name = "landscape"

# LARCH HIKE

[[rows]]
layout = "section"
title = "-25/10 north cascades"

[[rows]]
layout = "full"
[[rows.images]]
name = "larch-1"

# a comment inside the gallery

[[rows]]
layout = "full"
[[rows.images]]
name = "tall"
`
	want := `[metadata]
name = "landscape"
short_name = "landscape"

# LARCH HIKE

[[rows]]
layout = "section"
title = "-25/10 north cascades"

[[rows]]
layout = "full"
[[rows.images]]
name = "larch-1"
width = 2560
height = 1707

# a comment inside the gallery

[[rows]]
layout = "full"
[[rows.images]]
name = "tall"
width = 1707
height = 2560
`

	got, _ := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
}

func TestRecordDimensionsIgnoresNameKeysOutsideImageTables(t *testing.T) {
	content := `[metadata]
name = "larch-1"
`

	got, unknown := RecordDimensions(content, testDimensions)

	if got != content {
		t.Errorf("metadata name was modified:\n%s", got)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown = %v, want none", unknown)
	}
}

func TestRecordDimensionsHandlesHeroImages(t *testing.T) {
	content := `[[hero_images]]
name = "larch-1"

[[hero_images]]
name = "tall"
`
	want := `[[hero_images]]
name = "larch-1"
width = 2560
height = 1707

[[hero_images]]
name = "tall"
width = 1707
height = 2560
`

	got, _ := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
}

func TestRecordDimensionsReportsImagesWithoutAnOriginal(t *testing.T) {
	content := `[[rows.images]]
name = "larch-1"

[[rows.images]]
name = "not-exported"
`
	want := `[[rows.images]]
name = "larch-1"
width = 2560
height = 1707

[[rows.images]]
name = "not-exported"
`

	got, unknown := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
	if !slices.Equal(unknown, []string{"not-exported"}) {
		t.Errorf("unknown = %v, want [not-exported]", unknown)
	}
}

func TestRecordDimensionsHandlesFlexibleSpacing(t *testing.T) {
	content := `[[rows.images]]
name="larch-1"   # the first one
`
	want := `[[rows.images]]
name="larch-1"   # the first one
width = 2560
height = 1707
`

	got, _ := RecordDimensions(content, testDimensions)

	if got != want {
		t.Errorf("RecordDimensions =\n%s\nwant\n%s", got, want)
	}
}
