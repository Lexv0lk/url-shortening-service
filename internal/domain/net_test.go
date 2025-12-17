package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		url         string
		expectedErr error
	}

	testCases := []testCase{
		{
			name:        "valid https URL",
			url:         "https://example.com/path",
			expectedErr: nil,
		},
		{
			name:        "valid http URL",
			url:         "http://example.com",
			expectedErr: nil,
		},
		{
			name:        "unsupported scheme",
			url:         "ftp://example.com/file",
			expectedErr: &InvalidUrlError{},
		},
		{
			name:        "empty host",
			url:         "https://",
			expectedErr: &InvalidUrlError{},
		},
		{
			name:        "missing scheme",
			url:         "example.com",
			expectedErr: &InvalidUrlError{},
		},
		{
			name:        "completely invalid URL",
			url:         ":/invalid",
			expectedErr: &InvalidUrlError{},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateURL(tt.url)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
