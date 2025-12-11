package mocks

import "sync/atomic"

type SlowSafeIDGenerator struct {
	nextID *atomic.Uint64
}

func NewSlowSafeIDGenerator() *SlowSafeIDGenerator {
	return &SlowSafeIDGenerator{
		nextID: &atomic.Uint64{},
	}
}

func (g *SlowSafeIDGenerator) GetNextId() (uint64, error) {
	return g.nextID.Add(1), nil
}
