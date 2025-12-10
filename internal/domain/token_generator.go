package domain

import "strings"

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base     = uint64(len(alphabet))
)

// GenerateToken generates a short string token from a given uint64 ID
func GenerateToken(id uint64) string {
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

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
