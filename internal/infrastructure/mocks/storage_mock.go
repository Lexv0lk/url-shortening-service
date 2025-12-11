package mocks

import (
	"fmt"
	"url-shortening-service/internal/domain"
)

type LocalStorage struct {
	storage map[uint64]domain.MappingInfo
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		storage: make(map[uint64]domain.MappingInfo),
	}
}

func (s *LocalStorage) AddNewMapping(id uint64, originalUrl string, urlToken string) error {
	if _, found := s.storage[id]; found {
		return &domain.UrlExistingError{Msg: fmt.Sprintf("Duplicate mapping id %d", id)}
	}

	s.storage[id] = domain.MappingInfo{
		Id:          id,
		OriginalURL: originalUrl,
		Token:       urlToken,
	}

	return nil
}

func (s *LocalStorage) GetMapping(urlToken string) (domain.MappingInfo, bool) {
	for _, mapping := range s.storage {
		if mapping.Token == urlToken {
			return mapping, true
		}
	}

	return domain.MappingInfo{}, false
}

func (s *LocalStorage) GetLastId() (uint64, error) {
	var lastId uint64
	for id := range s.storage {
		if id > lastId {
			lastId = id
		}
	}

	return lastId, nil
}
