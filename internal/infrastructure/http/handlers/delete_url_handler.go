package handlers

import (
	"errors"
	"net/http"
	"url-shortening-service/internal/domain"
)

// DeleteUrlHandler handles HTTP requests for deleting URL mappings.
type DeleteUrlHandler struct {
	urlDeleter domain.UrlDeleter
	logger     domain.Logger
}

// NewDeleteUrlHandler creates a new DeleteUrlHandler instance.
// Parameters:
//   - urlDeleter: service for deleting URL mappings
//   - logger: logger for recording errors
func NewDeleteUrlHandler(urlDeleter domain.UrlDeleter, logger domain.Logger) *DeleteUrlHandler {
	return &DeleteUrlHandler{
		urlDeleter: urlDeleter,
		logger:     logger,
	}
}

// Delete handles DELETE requests to remove a URL mapping.
// It extracts the URL token from the path and deletes the corresponding mapping.
//
// HTTP Responses:
//   - 204 No Content: mapping successfully deleted
//   - 404 Not Found: URL token does not exist
//   - 500 Internal Server Error: unexpected error occurred
func (h *DeleteUrlHandler) Delete(w http.ResponseWriter, r *http.Request) {
	urlToken := r.PathValue(domain.UrlTokenStr)

	err := h.urlDeleter.DeleteUrl(r.Context(), urlToken)
	if errors.Is(err, &domain.TokenNonExistingError{}) {
		http.Error(w, "URL token not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error("Failed to delete URL: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
