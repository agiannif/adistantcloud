package handlers

import (
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

type Galleries struct {
	Galleries map[string]*config.GalleryConfig
}

func (g *Galleries) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	galleryShortName := r.PathValue("shortName")

	// temp default to "lost" until we add a home page and other galleries
	if galleryShortName == "" {
		galleryShortName = "lost"
	}

	galleryConfig, ok := g.Galleries[galleryShortName]
	if !ok {
		http.Error(w, "specified gallery not found", http.StatusBadRequest)
		return
	}

	template.Gallery(galleryConfig).Render(r.Context(), w)
}
