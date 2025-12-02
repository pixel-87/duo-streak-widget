package service

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"text/template"
	"time"
)

//go:embed templates/githubText.svg
var githubSVGTemplate string

// GithubService implements the api.Service interface.
type GithubService struct {
	tmpl    *template.Template
	client  *http.Client
	baseURL string

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
	return &GithubService{
		tmpl: tmpl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://github.com/users/%s/contributions",
		cache:   make(map[string]cacheEntry),
		ttl:     4 * time.Hour,
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

	// Fetch the contributions page
	url := fmt.Sprintf(s.baseURL, username)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Set headers for HTML response
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; StreakWidget/1.0)")
	req.Header.Set("Accept", "text/html")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github contributions page returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse contributions and calculate streak
	streak := s.calculateStreak(string(body))

	// Update cache
	s.mu.Lock()
	s.cache[username] = cacheEntry{
		streak:    streak,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return streak, nil
}

// calculateStreak parses GitHub contributions HTML to calculate the current streak.
// Looks for data-date and data-level attributes in the contribution graph.
func (s *GithubService) calculateStreak(html string) int {
	// GitHub contributions page contains elements like:
	// <td ... data-date="2025-12-01" data-level="1" ...>
	// data-level="0" means no contributions, 1-4 means contributions

	// Regex to find contribution days with dates and levels
	re := regexp.MustCompile(`data-date="(\d{4}-\d{2}-\d{2})"[^>]*data-level="(\d)"`)
	matches := re.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		// Try alternative pattern if order differs
		re = regexp.MustCompile(`data-level="(\d)"[^>]*data-date="(\d{4}-\d{2}-\d{2})"`)
		matches = re.FindAllStringSubmatch(html, -1)
		// Swap capture groups for consistency
		for i, m := range matches {
			if len(m) >= 3 {
				matches[i] = []string{m[0], m[2], m[1]}
			}
		}
	}

	if len(matches) == 0 {
		return 0
	}

	// Build map of date to contribution status
	contributions := make(map[string]bool)
	for _, match := range matches {
		if len(match) >= 3 {
			date := match[1]
			level, _ := strconv.Atoi(match[2])
			contributions[date] = level > 0
		}
	}

	// Calculate streak starting from today backwards
	streak := 0
	today := time.Now().UTC()

	// Iterate backwards from today
	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		hasContribution, exists := contributions[date]

		if i == 0 && !exists {
			// Skip today if not in data
			continue
		}

		if hasContribution {
			streak++
		} else if i > 0 {
			// Break streak if no contribution and not today
			break
		}
	}

	return streak
}
