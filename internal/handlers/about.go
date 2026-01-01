package handlers

import (
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

// About handles the about page
type About struct {
	Galleries []config.GalleryMetadata
}

// AboutHandler renders the about page
func (a *About) AboutHandler(w http.ResponseWriter, r *http.Request) {
	template.About(a.Galleries).Render(r.Context(), w)
}
