package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
)

type ShortenUrlHandler struct {
	urlAdder application.UrlShortener
	logger   domain.Logger
}

type ShortenUrlRequest struct {
	URL string `json:"url"`
}

type ShortenUrlResponse struct {
	ShortURL string `json:"short_url"`
}

func NewAddUrlHandler(urlAdder application.UrlShortener, logger domain.Logger) *ShortenUrlHandler {
	return &ShortenUrlHandler{
		urlAdder: urlAdder,
		logger:   logger,
	}
}

func (h *ShortenUrlHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ShortenUrlRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	urlToken, err := h.urlAdder.AddTokenForUrl(r.Context(), req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortUrl, err := url.JoinPath(domain.BaseUrl, urlToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ShortenUrlResponse{ShortURL: shortUrl}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
	}
}
