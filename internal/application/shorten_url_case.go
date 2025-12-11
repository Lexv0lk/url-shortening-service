package application

import (
	"context"
	"fmt"
	"net/url"
	"url-shortening-service/internal/domain"
)

var validSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

type UrlShortener struct {
	store       domain.MappingInfoAdder
	idGenerator domain.IdGenerator
}

func NewUrlShortener(idGenerator domain.IdGenerator, store domain.MappingInfoAdder) *UrlShortener {
	return &UrlShortener{
		store:       store,
		idGenerator: idGenerator,
	}
}

func (u *UrlShortener) AddTokenForUrl(ctx context.Context, originalUrl string) (string, error) {
	parsedUrl, err := url.ParseRequestURI(originalUrl)
	if err != nil || parsedUrl.Host == "" {
		return "", &domain.InvalidUrlError{Msg: fmt.Sprintf("Invalid url provided: %s", originalUrl)}
	} else if _, valid := validSchemes[parsedUrl.Scheme]; !valid {
		return "", &domain.InvalidUrlError{Msg: fmt.Sprintf("Unsupported URL scheme: %s", parsedUrl.Scheme)}
	}

	id, err := u.idGenerator.GetNextId(ctx)
	if err != nil {
		return "", err
	}

	urlToken := domain.GenerateToken(id)
	return urlToken, u.store.AddNewMapping(id, originalUrl, urlToken)
}
