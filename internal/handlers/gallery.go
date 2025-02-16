package handlers

import (
	"net/http"
	"strconv"

	"github.com/anybote/litehouse/internal/config"
	"github.com/anybote/litehouse/web/template"
)

type Gallery struct {
	GalleryConfig  config.GalleryConfig
	NumRowsPerPage int
}

func (g *Gallery) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	template.Gallery(&g.GalleryConfig, g.NumRowsPerPage).Render(r.Context(), w)
}

func (g *Gallery) ImagePageHandler(w http.ResponseWriter, r *http.Request) {
	numberOfPages := (len(g.GalleryConfig.Rows) + g.NumRowsPerPage - 1) / g.NumRowsPerPage
	indexPath := r.PathValue("index")
	index, err := strconv.Atoi(indexPath)
	if err != nil || index < 0 || index >= numberOfPages {
		http.Error(w, "Invalid page index", http.StatusBadRequest)
		return
	}

	template.ImagePage(index, &g.GalleryConfig, g.NumRowsPerPage).Render(r.Context(), w)
}
