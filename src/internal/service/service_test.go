package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchStreak(t *testing.T) {
	// 1. Create a fake Duolingo server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("User-Agent") != "curl/8.0.1" {
			t.Errorf("Expected User-Agent curl/8.0.1, got %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept application/json, got %s", r.Header.Get("Accept"))
		}

		// Return a fake JSON response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"users": [
				{
					"streak": 10,
					"streakData": {
						"currentStreak": {
							"length": 42
						}
					}
				}
			]
		}`)
	}))
	defer ts.Close()

	// 2. Create the service
	svc, err := NewDuoService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// 3. Override the base URL to point to our fake server
	svc.baseURL = ts.URL

	// 4. Call GetBadge (which calls fetchStreak)
	svgBytes, err := svc.GetBadge(context.Background(), "testuser", "default")
	if err != nil {
		t.Fatalf("GetBadge failed: %v", err)
	}

	// 5. Verify the output contains the streak number (42)
	svg := string(svgBytes)
	if !strings.Contains(svg, "42") {
		t.Errorf("Expected SVG to contain streak 42, got: %s", svg)
	}
}

func TestGetBadge_Caching(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"users":[{"streak":10,"streakData":{"currentStreak":{"length":100}}}]}`)
	}))
	defer ts.Close()

	svc, err := NewDuoService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	svc.baseURL = ts.URL

	// First call - should hit server
	_, err = svc.GetBadge(context.Background(), "cacheuser", "default")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("Expected 1 request, got %d", requestCount)
	}

	// Second call - should hit cache
	_, err = svc.GetBadge(context.Background(), "cacheuser", "default")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("Expected 1 request (cached), got %d", requestCount)
	}
}
