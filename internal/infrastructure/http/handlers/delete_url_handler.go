package handlers

import (
	"errors"
	"net/http"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
)

type DeleteUrlHandler struct {
	urlDeleter application.UrlDeleter
	logger     domain.Logger
}

func NewDeleteUrlHandler(urlDeleter application.UrlDeleter, logger domain.Logger) *DeleteUrlHandler {
	return &DeleteUrlHandler{
		urlDeleter: urlDeleter,
		logger:     logger,
	}
}

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
