package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pixel-87/duo-streak-widget/api"
	"github.com/pixel-87/duo-streak-widget/internal/service"
)

func main() {
	// Initialize services
	duoSvc, err := service.NewDuoService()
	if err != nil {
		log.Fatalf("Failed to initialize Duolingo service: %v", err)
	}

	githubSvc, err := service.NewGithubService()
	if err != nil {
		log.Fatalf("Failed to initialize GitHub service: %v", err)
	}

	// Initialize API handler
	// Inject services into API
	handler := api.NewAPI(duoSvc, githubSvc)

	// Set up router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/api/duolingo/button", handler.GetDuoButton)
	mux.HandleFunc("/api/github/button", handler.GetGithubButton)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
