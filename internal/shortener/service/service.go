package shortener

import (
	"context"
	"github.com/Dor1ma/url-shortener/internal/shortener/storage"
	pb "github.com/Dor1ma/url-shortener/pkg/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"net/url"
	"regexp"
	"time"
)

const maxUrlLength = 2048

var shortUrlRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{5,20}$`)

type Service struct {
	pb.UnimplementedUrlShortenerServer
	repo storage.Repository
}

func NewService(repo storage.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortUrl(ctx context.Context, req *pb.CreateShortUrlRequest) (*pb.CreateShortUrlResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "request was cancelled")
	default:
		originalUrl := req.GetOriginalUrl()

		if !isValidURL(originalUrl) {
			return nil, status.Error(codes.FailedPrecondition, "incorrect url")
		}

		shortUrl := generateShortUrl()

		err := s.repo.SaveUrl(originalUrl, shortUrl)
		if err != nil {
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
			return nil, status.Error(codes.NotFound, "short URL not found")
		}

		return &pb.GetOriginalUrlResponse{OriginalUrl: originalUrl}, nil
	}
}

func generateShortUrl() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	rand.Seed(time.Now().UnixNano())
	shortUrl := make([]byte, 10)
	for i := range shortUrl {
		shortUrl[i] = charset[rand.Intn(len(charset))]
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
