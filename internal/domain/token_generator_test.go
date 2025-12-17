package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		id          int64
		expectedRes string
	}

	testCases := []testCase{
		{
			name:        "ID 0",
			id:          0,
			expectedRes: "a",
		},
		{
			name:        "ID 1",
			id:          1,
			expectedRes: "b",
		},
		{
			name:        "ID 61",
			id:          61,
			expectedRes: "9",
		},
		{
			name:        "ID 543643",
			id:          543643,
			expectedRes: "crAB",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := GenerateToken(tt.id)
			assert.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestReverse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		input       string
		expectedRes string
	}

	testCases := []testCase{
		{
			name:        "empty string",
			input:       "",
			expectedRes: "",
		},
		{
			name:        "single character",
			input:       "a",
			expectedRes: "a",
		},
		{
			name:        "palindrome",
			input:       "madam",
			expectedRes: "madam",
		},
		{
			name:        "regular string",
			input:       "hello",
			expectedRes: "olleh",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := reverse(tt.input)
			assert.Equal(t, tt.expectedRes, res)
		})
	}
}
