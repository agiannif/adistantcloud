package handlers

import (
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

type Gallery struct {
	GalleryConfig config.GalleryConfig
}

func (g *Gallery) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	template.Gallery(&g.GalleryConfig).Render(r.Context(), w)
}
