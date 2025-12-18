package mocks

import (
	"context"
	"testing"
	"url-shortening-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_AddNewMapping(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		id             int64
		originalUrl    string
		urlToken       string
		setupStorage   map[int64]domain.MappingInfo
		expectedResult domain.MappingInfo
		expectedError  bool
	}

	testCases := []testCase{
		{
			name:         "Success - add new mapping",
			id:           1,
			originalUrl:  "https://example.com",
			urlToken:     "abc123",
			setupStorage: map[int64]domain.MappingInfo{},
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com",
				Token:       "abc123",
			},
			expectedError: false,
		},
		{
			name:        "Error - duplicate id",
			id:          1,
			originalUrl: "https://new.com",
			urlToken:    "xyz789",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://existing.com",
					Token:       "abc123",
				},
			},
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://existing.com",
				Token:       "abc123",
			},
			expectedError: true,
		},
		{
			name:        "Success - add mapping with different id",
			id:          2,
			originalUrl: "https://new.com",
			urlToken:    "xyz789",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://existing.com",
					Token:       "abc123",
				},
			},
			expectedResult: domain.MappingInfo{
				Id:          2,
				OriginalURL: "https://new.com",
				Token:       "xyz789",
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := NewLocalStorage()
			for id, info := range tt.setupStorage {
				storage.storage[id] = info
			}

			result, err := storage.AddNewMapping(context.Background(), tt.id, tt.originalUrl, tt.urlToken)

			if tt.expectedError {
				require.Error(t, err)
				var urlExistingErr *domain.UrlExistingError
				assert.ErrorAs(t, err, &urlExistingErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLocalStorage_GetMapping(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		urlToken       string
		setupStorage   map[int64]domain.MappingInfo
		expectedResult domain.MappingInfo
		expectedFound  bool
	}

	testCases := []testCase{
		{
			name:     "Success - mapping found",
			urlToken: "abc123",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://example.com",
					Token:       "abc123",
				},
			},
			expectedResult: domain.MappingInfo{
				Id:          1,
				OriginalURL: "https://example.com",
				Token:       "abc123",
			},
			expectedFound: true,
		},
		{
			name:           "Not found - empty storage",
			urlToken:       "abc123",
			setupStorage:   map[int64]domain.MappingInfo{},
			expectedResult: domain.MappingInfo{},
			expectedFound:  false,
		},
		{
			name:     "Not found - token does not exist",
			urlToken: "nonexistent",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://example.com",
					Token:       "abc123",
				},
			},
			expectedResult: domain.MappingInfo{},
			expectedFound:  false,
		},
		{
			name:     "Success - found among multiple mappings",
			urlToken: "xyz789",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://example1.com",
					Token:       "abc123",
				},
				2: {
					Id:          2,
					OriginalURL: "https://example2.com",
					Token:       "xyz789",
				},
				3: {
					Id:          3,
					OriginalURL: "https://example3.com",
					Token:       "def456",
				},
			},
			expectedResult: domain.MappingInfo{
				Id:          2,
				OriginalURL: "https://example2.com",
				Token:       "xyz789",
			},
			expectedFound: true,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := NewLocalStorage()
			for id, info := range tt.setupStorage {
				storage.storage[id] = info
			}

			result, found := storage.GetMapping(context.Background(), tt.urlToken)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLocalStorage_GetLastId(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		setupStorage map[int64]domain.MappingInfo
		expectedId   int64
	}

	testCases := []testCase{
		{
			name:         "Empty storage returns 0",
			setupStorage: map[int64]domain.MappingInfo{},
			expectedId:   0,
		},
		{
			name: "Single mapping returns its id",
			setupStorage: map[int64]domain.MappingInfo{
				5: {
					Id:          5,
					OriginalURL: "https://example.com",
					Token:       "abc123",
				},
			},
			expectedId: 5,
		},
		{
			name: "Multiple mappings returns highest id",
			setupStorage: map[int64]domain.MappingInfo{
				1: {
					Id:          1,
					OriginalURL: "https://example1.com",
					Token:       "abc123",
				},
				10: {
					Id:          10,
					OriginalURL: "https://example2.com",
					Token:       "xyz789",
				},
				5: {
					Id:          5,
					OriginalURL: "https://example3.com",
					Token:       "def456",
				},
			},
			expectedId: 10,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := NewLocalStorage()
			for id, info := range tt.setupStorage {
				storage.storage[id] = info
			}

			lastId, err := storage.GetLastId(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tt.expectedId, lastId)
		})
	}
}
