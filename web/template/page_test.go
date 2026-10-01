package template

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestPageDeclaresCharsetAndViewportInSeparateMetaTags(t *testing.T) {
	var page bytes.Buffer
	if err := Home(nil, nil).Render(context.Background(), &page); err != nil {
		t.Fatal(err)
	}
	html := page.String()

	for _, want := range []string{
		`<meta charset="UTF-8">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1.0">`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered page does not contain %s", want)
		}
	}

	// a meta element is either a charset declaration or named metadata, not both
	for _, tag := range regexp.MustCompile(`<meta [^>]*>`).FindAllString(html, -1) {
		if strings.Contains(tag, "charset") && strings.Contains(tag, "name=") {
			t.Errorf("meta tag combines charset and name: %s", tag)
		}
	}
}
