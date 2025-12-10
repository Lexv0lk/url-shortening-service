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
	baseURL     string
}

func NewUrlShortener(idGenerator IdGenerator, store MappingInfoAdder, baseURL string) *UrlShortener {
	return &UrlShortener{
		store:       store,
		baseURL:     baseURL,
		idGenerator: idGenerator,
	}
}

func (u *UrlShortener) ShortenNewUrl(originalUrl string) (string, error) {
	parsedUrl, err := url.Parse(originalUrl)
	if err != nil || parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		return "", &domain.InvalidUrlError{Msg: fmt.Sprintf("Invalid url provided: %s", originalUrl)}
	}

	id, err := u.idGenerator.GetNextId()
	if err != nil {
		return "", err
	}

	urlToken := domain.GenerateToken(id)
	shortUrl, err := url.JoinPath(u.baseURL, urlToken)
	if err != nil {
		return "", err
	}

	return shortUrl, u.store.AddNewMapping(id, originalUrl, urlToken)
}
