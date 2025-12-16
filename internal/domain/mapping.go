package domain

import "time"

// MappingInfo represents a URL mapping record with metadata.
// It contains the relationship between a short URL token and its original URL.
type MappingInfo struct {
	// Id is the unique identifier of the mapping record.
	Id int64 `json:"id"`
	// OriginalURL is the full original URL that was shortened.
	OriginalURL string `json:"original_url"`
	// Token is the short unique string used to identify this mapping.
	Token string `json:"url_token"`
	// CreatedAt is the timestamp when the mapping was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the mapping was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}
