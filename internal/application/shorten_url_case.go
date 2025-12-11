package application

import (
	"fmt"
	"net/url"
	"url-shortening-service/internal/domain"
)

type MappingInfoAdder interface {
	AddNewMapping(id uint64, originalUrl string, shortUrl string) error
}

type IdGenerator interface {
	GetNextId() (uint64, error)
}

type UrlShortener struct {
	store       MappingInfoAdder
	idGenerator IdGenerator
}

func NewUrlShortener(idGenerator IdGenerator, store MappingInfoAdder) *UrlShortener {
	return &UrlShortener{
		store:       store,
		idGenerator: idGenerator,
	}
}

func (u *UrlShortener) AddTokenForUrl(originalUrl string) (string, error) {
	parsedUrl, err := url.Parse(originalUrl)
	if err != nil || parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return "", &domain.InvalidUrlError{Msg: fmt.Sprintf("Invalid url provided: %s", originalUrl)}
	}

	id, err := u.idGenerator.GetNextId()
	if err != nil {
		return "", err
	}

	urlToken := domain.GenerateToken(id)
	return urlToken, u.store.AddNewMapping(id, originalUrl, urlToken)
}
