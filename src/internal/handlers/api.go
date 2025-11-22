package handlers

import (
	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Handlers(r *chi.Mux) {
	// Global middlewares
	r.Use(chimiddle.StripSlashes)

	r.Route("/api/duolingo", func(router chi.Router) {

		router.Get("/button", GetDuoButton)
	})
}
