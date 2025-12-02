package api

import (
	"fmt"
	"html"
	"net/http"
)

// buttonParams are the arguments passed for a specific button design.
type buttonParams struct {
	Username string
	Variant  string
}

// Error represents an error response for the SVG badge handlers.
type Error struct {
	// Error code
	Code int

	// Error message
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Code:    code,
		Message: message,
	}
	// Create a simple SVG error badge
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="360" height="32">
            <rect width="100%%" height="100%%" fill="#f8d7da"/>
            <text x="8" y="20" fill="#721c24" font-family="sans-serif" font-size="13">%s</text>
         </svg>`,
		html.EscapeString(resp.Message),
	)

	w.Header().Set("Content-Type", "image/svg+xml")
	// Allow CDN/public caching of error badges for a short period
	w.Header().Set("Cache-Control", "public, max-age=10800")
	// Ensure caches (CDNs) store variants based on Accept-Encoding
	w.Header().Add("Vary", "Accept-Encoding")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(svg))
}

// wrappers for error handling
var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, "Internal Server Error", http.StatusInternalServerError)
	}
)

// API holds the HTTP handler dependencies.
type API struct {
	duoSvc    Service
	githubSvc Service
}

// NewAPI creates a new API instance.
func NewAPI(duoSvc, githubSvc Service) *API {
	return &API{
		duoSvc:    duoSvc,
		githubSvc: githubSvc,
	}
}

// GetDuoButton handles the /api/duolingo/button route.
// Validates query params and delegates to the service.
func (a *API) GetDuoButton(w http.ResponseWriter, r *http.Request) {
	params := buttonParams{
		Username: r.URL.Query().Get("username"),
		Variant:  r.URL.Query().Get("variant"),
	}

	if params.Username == "" {
		writeError(w, "Missing 'username' parameter", http.StatusBadRequest)
		return
	}

	if params.Variant == "" {
		params.Variant = "default"
	}

	// Get badge from service
	svg, err := a.duoSvc.GetBadge(r.Context(), params.Username, params.Variant)
	if err != nil {
		// In production, check error type for appropriate status code.
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	// Cache SVG for 3 hours
	w.Header().Set("Cache-Control", "public, max-age=10800")
	// Vary on Accept-Encoding for cache variants
	w.Header().Add("Vary", "Accept-Encoding")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svg)
}

// GetGithubButton handles the /api/github/button route.
// Validates query params and delegates to the GitHub service.
func (a *API) GetGithubButton(w http.ResponseWriter, r *http.Request) {
	params := buttonParams{
		Username: r.URL.Query().Get("username"),
		Variant:  r.URL.Query().Get("variant"),
	}

	if params.Username == "" {
		writeError(w, "Missing 'username' parameter", http.StatusBadRequest)
		return
	}

	if params.Variant == "" {
		params.Variant = "default"
	}

	// Get badge from service
	svg, err := a.githubSvc.GetBadge(r.Context(), params.Username, params.Variant)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=10800")
	w.Header().Add("Vary", "Accept-Encoding")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svg)
}
