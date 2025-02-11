package storage

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
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

	mock.ExpectExec("INSERT INTO urls").
		WithArgs("https://example.com", "short1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveUrl("https://example.com", "short1")
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveUrl_WhenDbErrorPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectExec("INSERT INTO urls").
		WithArgs("https://example.com", "short1").
		WillReturnError(errors.New("db error"))

	err = repo.SaveUrl("https://example.com", "short1")
	assert.Error(t, err)
	assert.Equal(t, "failed to save URL: db error", err.Error())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUrlPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT original_url").
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

	mock.ExpectQuery("SELECT original_url").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	originalUrl, err := repo.GetUrl("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "short URL not found", err.Error())
	assert.Equal(t, "", originalUrl)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUrl_WhenDbErrorPostgres(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &PostgresRepository{db: db}

	mock.ExpectQuery("SELECT original_url").
		WithArgs("short1").
		WillReturnError(errors.New("db error"))

	originalUrl, err := repo.GetUrl("short1")
	assert.Error(t, err)
	assert.Equal(t, "failed to get URL: db error", err.Error())
	assert.Equal(t, "", originalUrl)

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
