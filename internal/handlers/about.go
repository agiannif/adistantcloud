package handlers

import (
	"net/http"

	"github.com/anybote/litehouse/web/template"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	template.About().Render(r.Context(), w)
}
