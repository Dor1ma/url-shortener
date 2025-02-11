package storage

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveUrl(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	originalUrl, err := repo.GetUrl("short1")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", originalUrl)

	shortUrl, err := repo.GetShortUrl("https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "short1", shortUrl)
}

func TestSaveUrl_WhenOriginalUrlAlreadyExists(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	err = repo.SaveUrl("https://example.com", "short2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original URL already exists: short1")
}

func TestSaveUrl_WhenShortUrlAlreadyExists(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example1.com", "short1")
	assert.NoError(t, err)

	err = repo.SaveUrl("https://example2.com", "short1")
	assert.Error(t, err)
	assert.Equal(t, "short URL already exists", err.Error())
}

func TestGetUrl_WhenUrlExists(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	originalUrl, err := repo.GetUrl("short1")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", originalUrl)
}

func TestGetShortUrl_WhenUrlExists(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	shortUrl, err := repo.GetShortUrl("https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "short1", shortUrl)
}

func TestGetUrl_WhenUrlDoesNotExist(t *testing.T) {
	repo := NewInMemoryRepository()

	originalUrl, err := repo.GetUrl("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "short URL not found", err.Error())
	assert.Empty(t, originalUrl)
}

func TestGetShortUrl_WhenOriginalUrlDoesNotExist(t *testing.T) {
	repo := NewInMemoryRepository()

	shortUrl, err := repo.GetShortUrl("https://nonexistent.com")
	assert.Error(t, err)
	assert.Equal(t, "original URL not found", err.Error())
	assert.Empty(t, shortUrl)
}

func TestConcurrency(t *testing.T) {
	repo := NewInMemoryRepository()

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := repo.SaveUrl("https://example.com/"+strconv.Itoa(i), "short"+strconv.Itoa(i))
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 1000; i++ {
		originalUrl, err := repo.GetUrl("short" + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/"+strconv.Itoa(i), originalUrl)

		shortUrl, err := repo.GetShortUrl("https://example.com/" + strconv.Itoa(i))
		assert.NoError(t, err)
		assert.Equal(t, "short"+strconv.Itoa(i), shortUrl)
	}
}
