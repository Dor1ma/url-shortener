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
	query := `INSERT INTO urls (original_url, short_url) VALUES ($1, $2)`
	_, err := r.db.Exec(query, originalUrl, shortUrl) // Используем Exec вместо ExecContext
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

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}
