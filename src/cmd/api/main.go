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
	// 1. Initialize the Service
	svc, err := service.NewDuoService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// 2. Initialize the API Handler
	// We inject the service into the API
	handler := api.NewAPI(svc)

	// 3. Set up the Router
	mux := http.NewServeMux()

	// Register the route
	mux.HandleFunc("/api/duolingo/button", handler.GetDuoButton)

	// 4. Start the Server
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
