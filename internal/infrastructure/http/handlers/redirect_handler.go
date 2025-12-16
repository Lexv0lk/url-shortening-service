package handlers

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
	"url-shortening-service/internal/application/urlcases"
	"url-shortening-service/internal/domain"
)

// RedirectHandler handles HTTP requests for URL redirection.
// It retrieves the original URL and redirects the client, while also
// sending statistics events for analytics.
type RedirectHandler struct {
	urlGetter   urlcases.UrlGetter
	statsSender domain.StatisticsSender
	logger      domain.Logger
}

type RedirectRequest struct {
	Token string
}

// NewRedirectHandler creates a new RedirectHandler instance.
// Parameters:
//   - urlGetter: service for retrieving original URLs
//   - statsSender: sender for statistics events
//   - logger: logger for recording warnings and errors
func NewRedirectHandler(urlGetter urlcases.UrlGetter, statsSender domain.StatisticsSender, logger domain.Logger) *RedirectHandler {
	return &RedirectHandler{
		urlGetter:   urlGetter,
		logger:      logger,
		statsSender: statsSender,
	}
}

// Redirect handles GET requests to redirect from short URL to original URL.
// It retrieves the original URL, sends a statistics event asynchronously,
// and redirects the client with HTTP 307 Temporary Redirect.
//
// HTTP Responses:
//   - 307 Temporary Redirect: successful redirect to original URL
//   - 404 Not Found: URL token does not exist
//   - 500 Internal Server Error: unexpected error occurred
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
		IP:        retrieveIP(r),
		UserAgent: r.UserAgent(),
		Referrer:  r.Referer(),
	})
	if err != nil {
		h.logger.Warn("Failed to send statistics event: " + err.Error())
	}

	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}

func retrieveIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		if ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0]); ip != "" {
			return ip
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
