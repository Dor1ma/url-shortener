package storage

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (Repository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) SaveUrl(originalUrl, shortUrl string) error {
	var existingShortUrl string
	queryCheck := `SELECT short_url FROM urls WHERE original_url = $1`
	err := r.db.QueryRow(queryCheck, originalUrl).Scan(&existingShortUrl)
	if err == nil {
		return fmt.Errorf("original URL already exists: %s", existingShortUrl)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check URL existence: %w", err)
	}

	queryInsert := `INSERT INTO urls (original_url, short_url) VALUES ($1, $2) ON CONFLICT (short_url) DO NOTHING`
	_, err = r.db.Exec(queryInsert, originalUrl, shortUrl)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetUrl(shortUrl string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	var originalUrl string
	err := r.db.QueryRow(query, shortUrl).Scan(&originalUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("short URL not found")
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}
	return originalUrl, nil
}

func (r *PostgresRepository) GetShortUrl(originalUrl string) (string, error) {
	query := `SELECT short_url FROM urls WHERE original_url = $1`
	var shortUrl string
	err := r.db.QueryRow(query, originalUrl).Scan(&shortUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("original URL not found")
		}
		return "", fmt.Errorf("failed to get short URL: %w", err)
	}
	return shortUrl, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
