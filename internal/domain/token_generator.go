package domain

import "strings"

const (
	// alphabet contains all valid characters for token generation (base62).
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// base is the number of characters in the alphabet (62).
	base = int64(len(alphabet))
)

// GenerateToken generates a short alphanumeric token from a given int64 ID.
// The token is a base62 encoded representation of the ID, using lowercase letters,
// uppercase letters, and digits.
//
// For ID = 0, it returns "a" (the first character in the alphabet).
// The generated token length grows logarithmically with the ID value.
//
// Example:
//
//	GenerateToken(0)  -> "a"
//	GenerateToken(1)  -> "b"
//	GenerateToken(62) -> "ba"
func GenerateToken(id int64) string {
	if id == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	for id > 0 {
		remainder := id % base
		sb.WriteByte(alphabet[remainder])
		id = id / base
	}

	return reverse(sb.String())
}

// reverse reverses a string and returns the result.
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
