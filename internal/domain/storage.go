package domain

import "context"

type MappedGetSetter interface {
	OriginalUrlGetter
	UrlTokenSetter
}

type OriginalUrlGetter interface {
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool)
}

type UrlTokenSetter interface {
	SetMapping(ctx context.Context, originalUrl, urlToken string) error
}

type MappingInfoGetter interface {
	GetMapping(urlToken string) (MappingInfo, bool)
}

type MappingInfoLastIdGetter interface {
	GetLastId() (uint64, error)
}

type MappingInfoAdder interface {
	AddNewMapping(id uint64, originalUrl string, shortUrl string) error
}

type IdGenerator interface {
	GetNextId(ctx context.Context) (uint64, error)
}
