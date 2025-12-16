package domain

import (
	"fmt"
	"net/url"
)

const (
	ShortenUrlAddress = "POST /shorten"
	// UrlTokenStr is the path parameter name for URL tokens.
	UrlTokenStr = "urlToken"
	// RedirectAddress is the route pattern for redirecting to original URLs.
	RedirectAddress = "GET /{" + UrlTokenStr + "}"
	// UpdateUrlAddress is the route pattern for updating existing URL mappings.
	UpdateUrlAddress = "PUT /{" + UrlTokenStr + "}"
	// DeleteUrlAddress is the route pattern for deleting URL mappings.
	DeleteUrlAddress = "DELETE /{" + UrlTokenStr + "}"
	// StatsUrlAddress is the route pattern for retrieving URL statistics.
	StatsUrlAddress = "GET /shorten/{" + UrlTokenStr + "}/stats"
)

var validSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// ValidateURL checks if the provided URL string is valid and has a supported scheme.
// It validates the URL format and ensures only HTTP or HTTPS schemes are accepted.
//
// Returns *InvalidUrlError if:
//   - The URL cannot be parsed
//   - The URL has an empty host
//   - The URL scheme is not http or https
//
// Returns nil if the URL is valid.
func ValidateURL(URL string) error {
	parsedUrl, err := url.ParseRequestURI(URL)

	if err != nil || parsedUrl.Host == "" {
		return &InvalidUrlError{Msg: fmt.Sprintf("Invalid url provided: %s", URL)}
	} else if _, valid := validSchemes[parsedUrl.Scheme]; !valid {
		return &InvalidUrlError{Msg: fmt.Sprintf("Unsupported URL scheme: %s", parsedUrl.Scheme)}
	}

	return nil
}
