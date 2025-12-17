package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresSettings_GetUrl(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		settings PostgresSettings
		expected string
	}

	testCases := []testCase{
		{
			name: "ssl disabled",
			settings: PostgresSettings{
				User:       "user",
				Password:   "pass",
				Host:       "localhost",
				Port:       "5432",
				DBName:     "shortener",
				SSlEnabled: false,
			},
			expected: "postgres://user:pass@localhost:5432/shortener?sslmode=disable",
		},
		{
			name: "ssl enabled",
			settings: PostgresSettings{
				User:       "user",
				Password:   "secret",
				Host:       "db.local",
				Port:       "6543",
				DBName:     "service",
				SSlEnabled: true,
			},
			expected: "postgres://user:secret@db.local:6543/service",
		},
		{
			name: "empty fields",
			settings: PostgresSettings{
				User:       "",
				Password:   "",
				Host:       "",
				Port:       "",
				DBName:     "",
				SSlEnabled: false,
			},
			expected: "postgres://:@:/?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tt.settings.GetUrl()
			assert.Equal(t, tt.expected, result)
		})
	}
}
