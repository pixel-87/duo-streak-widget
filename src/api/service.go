package api

import "context"

// Service defines the business logic for retrieving a badge.
// It is an interface so we can swap in a real implementation or a mock for testing.
type Service interface {
	GetBadge(ctx context.Context, username, variant string) ([]byte, error)
}
