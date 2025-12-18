package mocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalCache_SetMapping(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		originalUrl string
		urlToken    string
	}

	testCases := []testCase{
		{
			name:        "Success - set new mapping",
			originalUrl: "https://example.com",
			urlToken:    "abc123",
		},
		{
			name:        "Success - set mapping with empty token",
			originalUrl: "https://example.com",
			urlToken:    "",
		},
		{
			name:        "Success - set mapping with empty url",
			originalUrl: "",
			urlToken:    "xyz789",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := NewLocalCache()

			err := cache.SetMapping(context.Background(), tt.originalUrl, tt.urlToken)

			require.NoError(t, err)
			assert.Equal(t, tt.originalUrl, cache.storage[tt.urlToken])
		})
	}
}

func TestLocalCache_SetMapping_OverwriteExisting(t *testing.T) {
	t.Parallel()

	cache := NewLocalCache()
	cache.storage["abc123"] = "https://old.com"

	err := cache.SetMapping(context.Background(), "https://new.com", "abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://new.com", cache.storage["abc123"])
}

func TestLocalCache_GetOriginalUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		urlToken      string
		setupStorage  map[string]string
		expectedUrl   string
		expectedFound bool
	}

	testCases := []testCase{
		{
			name:     "Success - mapping found",
			urlToken: "abc123",
			setupStorage: map[string]string{
				"abc123": "https://example.com",
			},
			expectedUrl:   "https://example.com",
			expectedFound: true,
		},
		{
			name:          "Not found - empty storage",
			urlToken:      "abc123",
			setupStorage:  map[string]string{},
			expectedUrl:   "",
			expectedFound: false,
		},
		{
			name:     "Not found - token does not exist",
			urlToken: "nonexistent",
			setupStorage: map[string]string{
				"abc123": "https://example.com",
			},
			expectedUrl:   "",
			expectedFound: false,
		},
		{
			name:     "Success - found among multiple mappings",
			urlToken: "xyz789",
			setupStorage: map[string]string{
				"abc123": "https://example1.com",
				"xyz789": "https://example2.com",
				"def456": "https://example3.com",
			},
			expectedUrl:   "https://example2.com",
			expectedFound: true,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := NewLocalCache()
			for token, url := range tt.setupStorage {
				cache.storage[token] = url
			}

			url, found := cache.GetOriginalUrl(context.Background(), tt.urlToken)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedUrl, url)
		})
	}
}
