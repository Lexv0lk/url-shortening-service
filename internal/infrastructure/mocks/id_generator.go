package mocks

import (
	"context"
	"sync/atomic"
)

// SlowSafeIDGenerator is a thread-safe mock ID generator using atomic operations.
// It is intended for testing purposes only.
type SlowSafeIDGenerator struct {
	nextID *atomic.Uint64
}

// NewSlowSafeIDGenerator creates a new SlowSafeIDGenerator instance starting from 0.
func NewSlowSafeIDGenerator() *SlowSafeIDGenerator {
	return &SlowSafeIDGenerator{
		nextID: &atomic.Uint64{},
	}
}

// GetNextId atomically increments and returns the next unique ID.
// This method is thread-safe and always succeeds (never returns an error).
func (g *SlowSafeIDGenerator) GetNextId(ctx context.Context) (uint64, error) {
	return g.nextID.Add(1), nil
}
