package storage

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	repo := &PostgresRepository{db: db}

	err = repo.db.Ping()
	assert.NoError(t, err)
}

func TestSaveUrlPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	// Проверяем, что URL не существует
	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url =").
		WithArgs("https://example.com").
		WillReturnError(sql.ErrNoRows)

	// Вставляем новую запись
	mock.ExpectExec("INSERT INTO urls").
		WithArgs("https://example.com", "short1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveUrl_WhenOriginalUrlAlreadyExistsPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	// URL уже существует
	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url =").
		WithArgs("https://example.com").
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("short1"))

	err = repo.SaveUrl("https://example.com", "short2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original URL already exists: short1")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveUrl_WhenDbErrorPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url =").
		WithArgs("https://example.com").
		WillReturnError(errors.New("db error"))

	err = repo.SaveUrl("https://example.com", "short1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check URL existence: db error")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUrlPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT original_url FROM urls WHERE short_url =").
		WithArgs("short1").
		WillReturnRows(sqlmock.NewRows([]string{"original_url"}).AddRow("https://example.com"))

	originalUrl, err := repo.GetUrl("short1")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", originalUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUrl_WhenNotFoundPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT original_url FROM urls WHERE short_url =").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	originalUrl, err := repo.GetUrl("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "short URL not found", err.Error())
	assert.Empty(t, originalUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetShortUrlPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url =").
		WithArgs("https://example.com").
		WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("short1"))

	shortUrl, err := repo.GetShortUrl("https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "short1", shortUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetShortUrl_WhenNotFoundPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url =").
		WithArgs("https://nonexistent.com").
		WillReturnError(sql.ErrNoRows)

	shortUrl, err := repo.GetShortUrl("https://nonexistent.com")
	assert.Error(t, err)
	assert.Equal(t, "original URL not found", err.Error())
	assert.Empty(t, shortUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUrl_WhenDbErrorPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT original_url FROM urls WHERE short_url =").
		WithArgs("short1").
		WillReturnError(errors.New("db error"))

	originalUrl, err := repo.GetUrl("short1")
	assert.Error(t, err)
	assert.Equal(t, "failed to get URL: db error", err.Error())
	assert.Empty(t, originalUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClosePostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectClose()

	err = repo.Close()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
