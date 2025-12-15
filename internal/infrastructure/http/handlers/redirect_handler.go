package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
)

type RedirectHandler struct {
	urlGetter   urlcases.UrlGetter
	statsSender domain.StatisticsSender
	logger      domain.Logger
}

type RedirectRequest struct {
	Token string
}

func NewRedirectHandler(urlGetter urlcases.UrlGetter, statsSender domain.StatisticsSender, logger domain.Logger) *RedirectHandler {
	return &RedirectHandler{
		urlGetter:   urlGetter,
		logger:      logger,
		statsSender: statsSender,
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

	err = h.statsSender.SendEvent(context.Background(), domain.RawStatsEvent{
		UrlToken:  token,
		Timestamp: time.Now(),
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Referrer:  r.Referer(),
	})
	if err != nil {
		h.logger.Warn("Failed to send statistics event: " + err.Error())
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}
