package storage

import (
	"errors"
	"sync"
)

type InMemoryRepository struct {
	mu   sync.RWMutex
	urls map[string]string // shortUrl -> originalUrl
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		urls: make(map[string]string),
	}
}

func (r *InMemoryRepository) SaveUrl(originalUrl, shortUrl string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.urls[shortUrl]; exists {
		return errors.New("short URL already exists")
	}

	r.urls[shortUrl] = originalUrl
	return nil
}

func (r *InMemoryRepository) GetUrl(shortUrl string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	originalUrl, exists := r.urls[shortUrl]
	if !exists {
		return "", errors.New("short URL not found")
	}

	return originalUrl, nil
}

func (r *InMemoryRepository) Close() error {
	return nil
}
