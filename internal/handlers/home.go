package handlers

import (
	"math/rand/v2"
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

// Home handles the home page
type Home struct {
	HomeConfig config.HomeConfig
	Galleries  []config.GalleryMetadata
}

// HomeHandler renders the home page with a random hero image
func (h *Home) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Select random hero image
	var selectedHero string
	if len(h.HomeConfig.HeroImages) > 0 {
		selectedHero = h.HomeConfig.HeroImages[rand.IntN(len(h.HomeConfig.HeroImages))]
	}

	// Render template
	template.Home(selectedHero, h.Galleries).Render(r.Context(), w)
}
