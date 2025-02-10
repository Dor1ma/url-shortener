package storage

type Repository interface {
	SaveUrl(originalUrl, shortUrl string) error
	GetUrl(shortUrl string) (string, error)
	Close() error
}
