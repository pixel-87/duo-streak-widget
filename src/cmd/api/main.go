package main

import (
	"log"
	"net/http"
	"os"

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
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}