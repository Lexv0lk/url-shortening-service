package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
)

// UpdaterUrlHandler handles HTTP requests for updating URL mappings.
type UpdaterUrlHandler struct {
	urlUpdater urlcases.UrlUpdater
	logger     domain.Logger
}

type UpdateUrlRequest struct {
	NewURL string `json:"url"`
}

// NewUpdateUrlHandler creates a new UpdaterUrlHandler instance.
// Parameters:
//   - urlUpdater: service for updating URL mappings
//   - logger: logger for recording errors
func NewUpdateUrlHandler(urlUpdater urlcases.UrlUpdater, logger domain.Logger) *UpdaterUrlHandler {
	return &UpdaterUrlHandler{
		urlUpdater: urlUpdater,
		logger:     logger,
	}
}

// Update handles PUT requests to update an existing URL mapping.
// It expects a JSON body with the new URL and updates the mapping for the given token.
//
// HTTP Responses:
//   - 200 OK: URL successfully updated, returns updated MappingInfo JSON
//   - 400 Bad Request: invalid request payload or invalid URL format
//   - 404 Not Found: URL token does not exist
//   - 500 Internal Server Error: unexpected error occurred
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
