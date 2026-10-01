package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agiannif/adistantcloud/internal/config"
)

func TestHomeHandler(t *testing.T) {
	// Create home handler with test data
	homeConfig := config.HomeConfig{
		HeroImages: []config.ImageConfig{
			{Name: "test1", Width: 2560, Height: 1707},
			{Name: "test2", Width: 2560, Height: 1707},
		},
	}

	galleries := []config.GalleryMetadata{
		{Name: "Gallery One", ShortName: "one"},
		{Name: "Gallery Two", ShortName: "two"},
	}

	home := Home{
		HomeConfig: homeConfig,
		Galleries:  galleries,
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Call handler
	home.HomeHandler(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify one of the hero images appears in response
	body := w.Body.String()
	hasImage := strings.Contains(body, "/assets/images/test1-2000.avif") || strings.Contains(body, "/assets/images/test2-2000.avif")
	if !hasImage {
		t.Error("Expected response to contain one of the hero images")
	}
}
