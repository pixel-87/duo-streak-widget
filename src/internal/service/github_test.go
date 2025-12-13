package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGithubService_GetBadge_UsesGraphQLAndCaches(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	requestCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected Authorization Bearer test-token, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"data":{"user":{"contributionsCollection":{"contributionCalendar":{"weeks":[{"contributionDays":[{"date":"%s","contributionCount":1},{"date":"%s","contributionCount":1},{"date":"%s","contributionCount":0}]}]}}}}}`,
			today.Format("2006-01-02"),
			yesterday.Format("2006-01-02"),
			twoDaysAgo.Format("2006-01-02"),
		)
	}))
	defer ts.Close()

	svc, err := NewGithubService()
	if err != nil {
		t.Fatalf("NewGithubService failed: %v", err)
	}
	svc.graphqlURL = ts.URL

	svg, err := svc.GetBadge(context.Background(), "octocat", "default")
	if err != nil {
		t.Fatalf("GetBadge failed: %v", err)
	}
	if !strings.Contains(string(svg), "2") {
		t.Fatalf("expected streak 2 in svg, got: %s", string(svg))
	}

	_, err = svc.GetBadge(context.Background(), "octocat", "default")
	if err != nil {
		t.Fatalf("GetBadge second call failed: %v", err)
	}

	if requestCount != 1 {
		t.Fatalf("expected 1 API call due to cache, got %d", requestCount)
	}
}

func TestGithubService_AllowsMissingToken(t *testing.T) {
	// No token set; service should still initialize and omit Authorization header.
	today := time.Now().UTC()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("expected no Authorization header, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"data":{"user":{"contributionsCollection":{"contributionCalendar":{"weeks":[{"contributionDays":[{"date":"%s","contributionCount":1}]}]}}}}}`,
			today.Format("2006-01-02"),
		)
	}))
	defer ts.Close()

	svc, err := NewGithubService()
	if err != nil {
		t.Fatalf("NewGithubService failed without token: %v", err)
	}
	svc.graphqlURL = ts.URL

	_, err = svc.GetBadge(context.Background(), "octocat", "default")
	if err != nil {
		t.Fatalf("GetBadge failed without token: %v", err)
	}
}
