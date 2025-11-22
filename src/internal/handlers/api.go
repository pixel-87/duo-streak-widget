package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

// Handlers registers internal HTTP handlers on the provided router.
func Handlers(r *chi.Mux) {
	// Global middlewares
	r.Use(chimiddle.StripSlashes)

	r.Route("/api/duolingo", func(router chi.Router) {
		router.Get("/button", GetDuoButton)
	})
}

// GetDuoButton is a minimal placeholder handler to satisfy builds and
// CI linting. The real handler is implemented as a method in the
// `api` package and main.go wires that up directly. This placeholder
// returns 501 so that routes that import this package don't fail.
func GetDuoButton(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
