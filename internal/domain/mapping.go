package domain

import "time"

type MappingInfo struct {
	Id          int64     `json:"id"`
	OriginalURL string    `json:"original_url"`
	Token       string    `json:"url_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
