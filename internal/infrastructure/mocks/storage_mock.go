package mocks

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"
)

// LocalStorage is an in-memory mock implementation of URL mapping storage.
// It is intended for testing purposes only.
type LocalStorage struct {
	storage map[int64]domain.MappingInfo
}

// NewLocalStorage creates a new LocalStorage instance with an empty storage map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		storage: make(map[int64]domain.MappingInfo),
	}
}

// AddNewMapping adds a new URL mapping to the local storage.
// Returns *domain.UrlExistingError if a mapping with the given ID already exists.
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

// GetMapping retrieves a URL mapping by its token from the local storage.
// Returns the MappingInfo and true if found, or empty MappingInfo and false if not found.
func (s *LocalStorage) GetMapping(ctx context.Context, urlToken string) (domain.MappingInfo, bool) {
	for _, mapping := range s.storage {
		if mapping.Token == urlToken {
			return mapping, true
		}
	}

	return domain.MappingInfo{}, false
}

// GetLastId returns the highest ID currently in storage.
// Returns 0 if the storage is empty.
func (s *LocalStorage) GetLastId(ctx context.Context) (int64, error) {
	var lastId int64
	for id := range s.storage {
		if id > lastId {
			lastId = id
		}
	}

	return lastId, nil
}
