package mocks

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

type LocalStorage struct {
	storage map[int64]domain.MappingInfo
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		storage: make(map[int64]domain.MappingInfo),
	}
}

func (s *LocalStorage) AddNewMapping(ctx context.Context, id int64, originalUrl string, urlToken string) (domain.MappingInfo, error) {
	if info, found := s.storage[id]; found {
		return info, &domain.UrlExistingError{Msg: fmt.Sprintf("Duplicate mapping id %d", id)}
	}

	s.storage[id] = domain.MappingInfo{
		Id:          id,
		OriginalURL: originalUrl,
		Token:       urlToken,
	}

	return s.storage[id], nil
}

func (s *LocalStorage) GetMapping(ctx context.Context, urlToken string) (domain.MappingInfo, bool) {
	for _, mapping := range s.storage {
		if mapping.Token == urlToken {
			return mapping, true
		}
	}

	return domain.MappingInfo{}, false
}

func (s *LocalStorage) GetLastId(ctx context.Context) (int64, error) {
	var lastId int64
	for id := range s.storage {
		if id > lastId {
			lastId = id
		}
	}

	return lastId, nil
}
