package handlers

import (
	"net/http"
	"strconv"

	"github.com/agiannif/adistantcloud/internal/config"
	"github.com/agiannif/adistantcloud/web/template"
)

type Gallery struct {
	GalleryConfig  config.GalleryConfig
	NumRowsPerPage int
}

func (g *Gallery) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	sectionVal := r.PathValue("section")
	if sectionVal == "" {
		// default to the first section if none is specified
		sectionVal = "0"
	}
	section, err := strconv.Atoi(sectionVal)
	if err != nil || section < 0 {
		http.Error(w, "Invalid gallery section", http.StatusBadRequest)
	}

	template.Gallery(&g.GalleryConfig, g.NumRowsPerPage, section).Render(r.Context(), w)
}

func (g *Gallery) ImagePageHandler(w http.ResponseWriter, r *http.Request) {
	indexVal := r.PathValue("index")
	index, err := strconv.Atoi(indexVal)
	if err != nil || index < 0 || index >= len(g.GalleryConfig.Rows) {
		http.Error(w, "Invalid row index", http.StatusBadRequest)
		return
	}

	template.ImagePage(&g.GalleryConfig, g.NumRowsPerPage, index).Render(r.Context(), w)
}
