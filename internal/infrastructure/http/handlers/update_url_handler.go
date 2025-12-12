package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
)

type UpdaterUrlHandler struct {
	urlUpdater application.UrlUpdater
	logger     domain.Logger
}

type UpdateUrlRequest struct {
	NewURL string `json:"url"`
}

func NewUpdateUrlHandler(urlUpdater application.UrlUpdater, logger domain.Logger) *UpdaterUrlHandler {
	return &UpdaterUrlHandler{
		urlUpdater: urlUpdater,
		logger:     logger,
	}
}

func (h *UpdaterUrlHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateUrlRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token := r.PathValue(domain.UrlTokenStr)

	mappingInfo, err := h.urlUpdater.UpdateUrlMapping(r.Context(), token, req.NewURL)
	if errors.Is(err, &domain.InvalidUrlError{}) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if errors.Is(err, &domain.TokenNonExistingError{}) {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to update URL mapping: %v", err))
		http.Error(w, "Server internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(mappingInfo)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
	}
}
