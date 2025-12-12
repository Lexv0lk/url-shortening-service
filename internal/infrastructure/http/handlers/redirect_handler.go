package handlers

import (
	"errors"
	"net/http"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
)

type RedirectHandler struct {
	urlGetter application.UrlGetter
	logger    domain.Logger
}

type RedirectRequest struct {
	Token string
}

func NewRedirectHandler(urlGetter application.UrlGetter, logger domain.Logger) *RedirectHandler {
	return &RedirectHandler{
		urlGetter: urlGetter,
		logger:    logger,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue(domain.UrlTokenStr)

	originalUrl, err := h.urlGetter.GetOriginalUrl(r.Context(), token)
	if errors.Is(err, &domain.UrlNonExistingError{}) {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Error("Failed to get original URL: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}
