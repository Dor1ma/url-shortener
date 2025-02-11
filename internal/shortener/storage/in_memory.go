package storage

import (
	"errors"
	"sync"
)

type InMemoryRepository struct {
	mu              sync.RWMutex
	shortToOriginal map[string]string
	originalToShort map[string]string
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

func (r *InMemoryRepository) SaveUrl(originalUrl, shortUrl string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShortUrl, exists := r.originalToShort[originalUrl]; exists {
		return errors.New("original URL already exists: " + existingShortUrl)
	}

	if _, exists := r.shortToOriginal[shortUrl]; exists {
		return errors.New("short URL already exists")
	}

	r.shortToOriginal[shortUrl] = originalUrl
	r.originalToShort[originalUrl] = shortUrl
	return nil
}

func (r *InMemoryRepository) GetUrl(shortUrl string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	originalUrl, exists := r.shortToOriginal[shortUrl]
	if !exists {
		return "", errors.New("short URL not found")
	}

	return originalUrl, nil
}

func (r *InMemoryRepository) GetShortUrl(originalUrl string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortUrl, exists := r.originalToShort[originalUrl]
	if !exists {
		return "", errors.New("original URL not found")
	}

	return shortUrl, nil
}

func (r *InMemoryRepository) Close() error {
	return nil
}
