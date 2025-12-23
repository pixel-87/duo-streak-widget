package service

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"text/template"
	"time"
)

//go:embed templates/duoText.svg
var duoSVGTemplate string

type duoResponse struct {
	Users []duoUser `json:"users"`
}

type duoUser struct {
	Streak     int           `json:"streak"`
	StreakData duoStreakData `json:"streakData"`
}

type duoStreakData struct {
	CurrentStreak  duoCurrentStreak `json:"currentStreak"`
	PreviousStreak interface{}      `json:"previousStreak"`
}

type duoCurrentStreak struct {
	Length int `json:"length"`
}

type cacheEntry struct {
	streak    int
	expiresAt time.Time
}

// DuoService implements the api.Service interface.
type DuoService struct {
	tmpl    *template.Template
	client  *http.Client
	baseURL string

	cache map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewDuoService creates a new instance of DuoService.
func NewDuoService() (*DuoService, error) {
	tmpl, err := template.New("duo").Parse(duoSVGTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG template: %w", err)
	}
	return &DuoService{
		tmpl: tmpl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://www.duolingo.com/2017-06-30/users",
		cache:   make(map[string]cacheEntry),
		ttl:     4 * time.Hour,
	}, nil
}

// GetBadge fetches the streak and renders the badge.
func (s *DuoService) GetBadge(ctx context.Context, username, variant string) ([]byte, error) {
	// 1. Fetch streak from Duolingo
	streak, err := s.fetchStreak(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch streak: %w", err)
	}

	// 2. Render the SVG
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

func (s *DuoService) fetchStreak(ctx context.Context, username string) (int, error) {
	// Check cache
	s.mu.RLock()
	entry, ok := s.cache[username]
	s.mu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.streak, nil
	}

	// Construct the URL with query parameters
	u, err := url.Parse(s.baseURL)
	if err != nil {
		return 0, err
	}
	q := u.Query()
	q.Set("username", username)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return 0, err
	}

	// Set User-Agent to avoid being blocked
	// Mirror the curl header used in tests to ensure Duolingo returns JSON
	req.Header.Set("User-Agent", "curl/8.0.1")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("duolingo api returned status: %d", resp.StatusCode)
	}

	var duoResp duoResponse
	if err := json.NewDecoder(resp.Body).Decode(&duoResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(duoResp.Users) == 0 {
		return 0, fmt.Errorf("user not found")
	}

	user := duoResp.Users[0]
	streak := user.Streak
	// Prefer the explicit streakData length when available (matches JS example)
	if user.StreakData.CurrentStreak.Length != 0 {
		streak = user.StreakData.CurrentStreak.Length
	}

	// Update cache
	s.mu.Lock()
	s.cache[username] = cacheEntry{
		streak:    streak,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return streak, nil
}
