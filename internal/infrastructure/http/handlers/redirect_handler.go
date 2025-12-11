package handlers

import (
	"net/http"
	"url-shortening-service/internal/application"
	"url-shortening-service/internal/domain"
)

type RedirectHandler struct {
	urlGetter application.UrlGetter
}

type RedirectRequest struct {
	Token string
}

func NewRedirectHandler(urlGetter application.UrlGetter) *RedirectHandler {
	return &RedirectHandler{
		urlGetter: urlGetter,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue(domain.UrlTokenStr)

	originalUrl, err := h.urlGetter.GetOriginalUrl(token)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}
