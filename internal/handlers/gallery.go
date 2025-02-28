package handlers

import (
	"net/http"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

type Gallery struct {
	GalleryConfig  config.GalleryConfig
	NumRowsPerPage int
}

func (g *Gallery) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	template.Gallery(&g.GalleryConfig, g.NumRowsPerPage).Render(r.Context(), w)
}
