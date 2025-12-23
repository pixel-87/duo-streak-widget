package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"
)

//go:embed templates/githubText.svg
var githubSVGTemplate string

const githubGraphQLQuery = `query($login: String!) {
	user(login: $login) {
		contributionsCollection {
			contributionCalendar {
				weeks {
					contributionDays {
						date
						contributionCount
					}
				}
			}
		}
	}
}`

type githubGraphQLRequest struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

type githubGraphQLResponse struct {
	Data   githubGraphQLData    `json:"data"`
	Errors []githubGraphQLError `json:"errors"`
}

type githubGraphQLData struct {
	User githubGraphQLUser `json:"user"`
}

type githubGraphQLUser struct {
	Contributions githubContributions `json:"contributionsCollection"`
}

type githubContributions struct {
	Calendar githubCalendar `json:"contributionCalendar"`
}

type githubCalendar struct {
	Weeks []githubWeek `json:"weeks"`
}

type githubWeek struct {
	Days []githubDay `json:"contributionDays"`
}

type githubDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}

type githubGraphQLError struct {
	Message string `json:"message"`
}

// GithubService implements the api.Service interface.
type GithubService struct {
	tmpl       *template.Template
	client     *http.Client
	graphqlURL string
	token      string

	cache map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewGithubService creates a new GithubService.
func NewGithubService() (*GithubService, error) {
	tmpl, err := template.New("github").Parse(githubSVGTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GitHub SVG template: %w", err)
	}
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	return &GithubService{
		tmpl: tmpl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		graphqlURL: "https://api.github.com/graphql",
		token:      token,
		cache:      make(map[string]cacheEntry),
		ttl:        4 * time.Hour,
	}, nil
}

// GetBadge fetches the GitHub contribution streak and renders the badge.
func (s *GithubService) GetBadge(ctx context.Context, username, variant string) ([]byte, error) {
	// Fetch streak from GitHub contributions page
	streak, err := s.fetchStreak(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub streak: %w", err)
	}

	// Render the SVG
	var buf bytes.Buffer
	data := struct {
		Streak int
	}{
		Streak: streak,
	}

	if err := s.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render SVG: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *GithubService) fetchStreak(ctx context.Context, username string) (int, error) {
	// Check cache first
	s.mu.RLock()
	entry, ok := s.cache[username]
	s.mu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.streak, nil
	}

	payload := githubGraphQLRequest{
		Query: githubGraphQLQuery,
		Variables: map[string]string{
			"login": username,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to encode GraphQL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.graphqlURL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "StreakWidget/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized {
		return 0, fmt.Errorf("github graphql api returned 401 unauthorized; provide GITHUB_TOKEN to increase limits and access")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github graphql api returned status: %d", resp.StatusCode)
	}

	var gqlResp githubGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return 0, fmt.Errorf("failed to decode github response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return 0, fmt.Errorf("github graphql error: %s", gqlResp.Errors[0].Message)
	}

	weeks := gqlResp.Data.User.Contributions.Calendar.Weeks
	if len(weeks) == 0 {
		return 0, fmt.Errorf("no contribution data returned for user: %s", username)
	}

	contributions := make(map[string]bool)
	for _, week := range weeks {
		for _, day := range week.Days {
			contributions[day.Date] = day.ContributionCount > 0
		}
	}

	streak := s.calculateStreak(contributions)

	// Update cache
	s.mu.Lock()
	s.cache[username] = cacheEntry{
		streak:    streak,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return streak, nil
}

// calculateStreak computes the current contribution streak from a date->bool map.
func (s *GithubService) calculateStreak(contributions map[string]bool) int {
	if len(contributions) == 0 {
		return 0
	}

	streak := 0
	today := time.Now().UTC()

	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		hasContribution, exists := contributions[date]

		if i == 0 && !exists {
			// Skip today if not in returned data (possible time zone gap)
			continue
		}

		if hasContribution {
			streak++
		} else if i > 0 {
			break
		}
	}

	return streak
}
