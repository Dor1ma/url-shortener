package shortener

import (
	"context"
	"math/rand"
	"net/url"
	"regexp"
	"sync"
	"time"

	pb "github.com/Dor1ma/url-shortener/api/gen/go"
	"github.com/Dor1ma/url-shortener/internal/shortener/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxUrlLength   = 2048
	shortUrlLength = 10
	charset        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

var (
	shortUrlRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{5,20}$`)
	rng           = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu            sync.Mutex
)

type Service struct {
	pb.UnimplementedUrlShortenerServer
	repo   storage.Repository
	logger *logrus.Logger
}

func NewService(repo storage.Repository, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateShortUrl(ctx context.Context, req *pb.CreateShortUrlRequest) (*pb.CreateShortUrlResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request was cancelled")
	default:
		originalUrl := req.GetOriginalUrl()

		if len(originalUrl) > maxUrlLength {
			return nil, status.Error(codes.FailedPrecondition, "URL exceeds maximum length")
		}

		if !isValidURL(originalUrl) {
			return nil, status.Error(codes.FailedPrecondition, "incorrect url")
		}

		existingShortUrl, err := s.repo.GetShortUrl(originalUrl)
		if err == nil {
			return &pb.CreateShortUrlResponse{ShortUrl: existingShortUrl}, nil
		}

		shortUrl := s.generateUniqueShortUrl()

		err = s.repo.SaveUrl(originalUrl, shortUrl)
		if err != nil {
			s.logger.Errorf("Error in repository occurred: %v", err)
			return nil, status.Error(codes.Internal, "failed to save URL")
		}

		return &pb.CreateShortUrlResponse{ShortUrl: shortUrl}, nil
	}
}

func (s *Service) GetOriginalUrl(ctx context.Context, req *pb.GetOriginalUrlRequest) (*pb.GetOriginalUrlResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request was canceled")
	default:
		shortUrl := req.GetShortUrl()

		if !isValidShortUrl(shortUrl) {
			return nil, status.Error(codes.FailedPrecondition, "incorrect url")
		}

		originalUrl, err := s.repo.GetUrl(shortUrl)
		if err != nil {
			s.logger.Errorf("Error in repository occurred: %v", err)
			return nil, status.Error(codes.NotFound, "short URL not found")
		}

		return &pb.GetOriginalUrlResponse{OriginalUrl: originalUrl}, nil
	}
}

func (s *Service) generateUniqueShortUrl() string {
	for {
		shortUrl := generateShortUrl()
		if _, err := s.repo.GetUrl(shortUrl); err != nil {
			return shortUrl
		}
	}
}

func generateShortUrl() string {
	mu.Lock()
	defer mu.Unlock()

	shortUrl := make([]byte, shortUrlLength)
	for i := range shortUrl {
		shortUrl[i] = charset[rng.Intn(len(charset))]
	}
	return string(shortUrl)
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isValidShortUrl(shortUrl string) bool {
	return shortUrlRegex.MatchString(shortUrl)
}
