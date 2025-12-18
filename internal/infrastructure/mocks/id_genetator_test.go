package mocks

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlowSafeIDGenerator_GetNextId(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		callsCount int
		expectedId uint64
	}

	testCases := []testCase{
		{
			name:       "First call returns 1",
			callsCount: 1,
			expectedId: 1,
		},
		{
			name:       "Second call returns 2",
			callsCount: 2,
			expectedId: 2,
		},
		{
			name:       "Multiple calls increment correctly",
			callsCount: 5,
			expectedId: 5,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			generator := NewSlowSafeIDGenerator()

			var id uint64
			var err error
			for i := 0; i < tt.callsCount; i++ {
				id, err = generator.GetNextId(context.Background())
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedId, id)
		})
	}
}

func TestSlowSafeIDGenerator_GetNextId_Concurrent(t *testing.T) {
	t.Parallel()

	generator := NewSlowSafeIDGenerator()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make(chan uint64, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			id, err := generator.GetNextId(context.Background())
			require.NoError(t, err)
			results <- id
		}()
	}

	wg.Wait()
	close(results)

	seen := make(map[uint64]bool)
	for id := range results {
		assert.False(t, seen[id], "Duplicate ID generated: %d", id)
		seen[id] = true
	}

	assert.Len(t, seen, goroutines)
	assert.Equal(t, uint64(goroutines), generator.nextID.Load())
}
