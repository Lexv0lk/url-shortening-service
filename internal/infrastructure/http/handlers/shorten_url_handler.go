package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
)

type ShortenUrlHandler struct {
	urlShortener urlcases.UrlShortener
	logger       domain.Logger
}

type ShortenUrlRequest struct {
	URL string `json:"url"`
}

func NewAddUrlHandler(urlShortener urlcases.UrlShortener, logger domain.Logger) *ShortenUrlHandler {
	return &ShortenUrlHandler{
		urlShortener: urlShortener,
		logger:       logger,
	}
}

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
