package storage

import (
	_ "errors"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestSaveUrl(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	originalUrl, err := repo.GetUrl("short1")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", originalUrl)
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

func TestGetUrl_WhenUrlDoesNotExist(t *testing.T) {
	repo := NewInMemoryRepository()

	originalUrl, err := repo.GetUrl("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "short URL not found", err.Error())
	assert.Equal(t, "", originalUrl)
}

func TestConcurrency(t *testing.T) {
	repo := NewInMemoryRepository()

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := repo.SaveUrl("https://example.com", "short"+strconv.Itoa(i))
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 1000; i++ {
		_, err := repo.GetUrl("short" + strconv.Itoa(i))
		assert.NoError(t, err)
	}
}
