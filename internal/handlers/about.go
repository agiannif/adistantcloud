package handlers

import (
	"net/http"

	"github.com/agiannif/adistantcloud/web/template"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	template.About().Render(r.Context(), w)
}
