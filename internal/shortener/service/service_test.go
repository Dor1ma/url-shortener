package shortener

import (
	"context"
	"errors"
	"testing"

	pb "github.com/Dor1ma/url-shortener/api/gen/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveUrl(originalUrl, shortUrl string) error {
	args := m.Called(originalUrl, shortUrl)
	return args.Error(0)
}

func (m *MockRepository) GetUrl(shortUrl string) (string, error) {
	args := m.Called(shortUrl)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetShortUrl(originalUrl string) (string, error) {
	args := m.Called(originalUrl)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestService_CreateShortUrl_ContextCancellation(t *testing.T) {
	repo := new(MockRepository)
	logger := logrus.New()
	service := NewService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &pb.CreateShortUrlRequest{OriginalUrl: "https://example.com"}
	_, err := service.CreateShortUrl(ctx, req)

	if err == nil || status.Code(err) != codes.Canceled {
		t.Errorf("CreateShortUrl() expected canceled context error, got: %v", err)
	}
}

func TestService_CreateShortUrl_ExistingShortUrl(t *testing.T) {
	repo := new(MockRepository)
	logger := logrus.New()
	service := NewService(repo, logger)

	repo.On("GetShortUrl", "https://example.com").Return("short123", nil)

	req := &pb.CreateShortUrlRequest{OriginalUrl: "https://example.com"}
	resp, err := service.CreateShortUrl(context.Background(), req)

	if err != nil {
		t.Errorf("CreateShortUrl() expected no error, got: %v", err)
	}

	if resp.ShortUrl != "short123" {
		t.Errorf("CreateShortUrl() expected short123, got: %v", resp.ShortUrl)
	}

	repo.AssertExpectations(t)
}

func TestService_CreateShortUrl_GenerateNewShortUrl(t *testing.T) {
	repo := new(MockRepository)
	logger := logrus.New()
	service := NewService(repo, logger)

	repo.On("GetShortUrl", "https://example.com").Return("", errors.New("not found"))
	repo.On("GetUrl", mock.Anything).Return("", errors.New("not found"))
	repo.On("SaveUrl", "https://example.com", mock.Anything).Return(nil)

	req := &pb.CreateShortUrlRequest{OriginalUrl: "https://example.com"}
	resp, err := service.CreateShortUrl(context.Background(), req)

	if err != nil {
		t.Errorf("CreateShortUrl() expected no error, got: %v", err)
	}

	if resp.ShortUrl == "" {
		t.Error("CreateShortUrl() expected a non-empty short URL")
	}

	repo.AssertExpectations(t)
}

func TestService_GetOriginalUrl_ContextCancellation(t *testing.T) {
	repo := new(MockRepository)
	logger := logrus.New()
	service := NewService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &pb.GetOriginalUrlRequest{ShortUrl: "validShortUrl"}
	_, err := service.GetOriginalUrl(ctx, req)

	if err == nil || status.Code(err) != codes.Canceled {
		t.Errorf("GetOriginalUrl() expected canceled context error, got: %v", err)
	}
}

func TestService_GetOriginalUrl_RepoError(t *testing.T) {
	repo := new(MockRepository)
	logger := logrus.New()
	service := NewService(repo, logger)

	repo.On("GetUrl", "nonexistent").Return("", errors.New("short URL not found"))

	req := &pb.GetOriginalUrlRequest{ShortUrl: "nonexistent"}
	_, err := service.GetOriginalUrl(context.Background(), req)

	if err == nil || status.Code(err) != codes.NotFound {
		t.Errorf("GetOriginalUrl() expected NotFound error, got: %v", err)
	}

	repo.AssertExpectations(t)
}
