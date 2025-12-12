package domain

import (
	"fmt"
	"net/url"
)

const (
	BaseUrl = `http://localhost:8080`
)

const (
	ShortenUrlAddress = "POST /shorten"
	UrlTokenStr       = "urlToken"
	RedirectAddress   = "GET /{" + UrlTokenStr + "}"
	UpdateUrlAddress  = "PUT /{" + UrlTokenStr + "}"
)

var validSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

func ValidateURL(URL string) error {
	parsedUrl, err := url.ParseRequestURI(URL)

	if err != nil || parsedUrl.Host == "" {
		return &InvalidUrlError{Msg: fmt.Sprintf("Invalid url provided: %s", URL)}
	} else if _, valid := validSchemes[parsedUrl.Scheme]; !valid {
		return &InvalidUrlError{Msg: fmt.Sprintf("Unsupported URL scheme: %s", parsedUrl.Scheme)}
	}

	return nil
}
