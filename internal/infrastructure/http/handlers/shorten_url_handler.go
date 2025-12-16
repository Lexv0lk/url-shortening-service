package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
)

// ShortenUrlHandler handles HTTP requests for creating shortened URLs.
type ShortenUrlHandler struct {
	urlShortener urlcases.UrlShortener
	logger       domain.Logger
}

type ShortenUrlRequest struct {
	URL string `json:"url"`
}

// NewAddUrlHandler creates a new ShortenUrlHandler instance.
// Parameters:
//   - urlShortener: service for creating shortened URLs
//   - logger: logger for recording errors
func NewAddUrlHandler(urlShortener urlcases.UrlShortener, logger domain.Logger) *ShortenUrlHandler {
	return &ShortenUrlHandler{
		urlShortener: urlShortener,
		logger:       logger,
	}
}

// Create handles POST requests to create a new shortened URL.
// It expects a JSON body with the original URL and returns the created mapping.
//
// HTTP Responses:
//   - 201 Created: URL successfully shortened, returns MappingInfo JSON
//   - 400 Bad Request: invalid request payload or invalid URL format
//   - 500 Internal Server Error: unexpected error occurred
func (h *ShortenUrlHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ShortenUrlRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mappingInfo, err := h.urlShortener.ShortenUrl(r.Context(), req.URL)
	if errors.Is(err, &domain.InvalidUrlError{}) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	} else if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to shorten URL: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(mappingInfo)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
	}
}
