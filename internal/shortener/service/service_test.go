package shortener

import (
	"context"
	"errors"
	pb "github.com/Dor1ma/url-shortener/api/gen/go"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
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

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestService_CreateShortUrl_ContextCancellation(t *testing.T) {
	repo := new(MockRepository)
	service := NewService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &pb.CreateShortUrlRequest{OriginalUrl: "https://example.com"}
	_, err := service.CreateShortUrl(ctx, req)

	if err == nil || status.Code(err) != codes.Canceled {
		t.Errorf("CreateShortUrl() expected canceled context error, got: %v", err)
	}
}

func TestService_CreateShortUrl_MaxURLLength(t *testing.T) {
	repo := new(MockRepository)
	service := NewService(repo)

	longURL := "https://example.com/" + string(make([]byte, maxUrlLength+1))
	req := &pb.CreateShortUrlRequest{OriginalUrl: longURL}

	_, err := service.CreateShortUrl(context.Background(), req)
	if err == nil {
		t.Error("CreateShortUrl() expected error for URL exceeding max length, got nil")
	}

	if status.Code(err) != codes.FailedPrecondition {
		t.Errorf("CreateShortUrl() expected FailedPrecondition error, got: %v", err)
	}
}

func TestService_CreateShortUrl_InvalidLength(t *testing.T) {
	repo := new(MockRepository)
	service := NewService(repo)

	longURL := "https://example.com/" + string(make([]byte, maxUrlLength+1))
	req := &pb.CreateShortUrlRequest{OriginalUrl: longURL}

	_, err := service.CreateShortUrl(context.Background(), req)
	if err == nil || status.Code(err) != codes.FailedPrecondition {
		t.Errorf("CreateShortUrl() expected FailedPrecondition error, got: %v", err)
	}
}

func TestService_GetOriginalUrl_ContextCancellation(t *testing.T) {
	repo := new(MockRepository)
	service := NewService(repo)

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
	service := NewService(repo)

	req := &pb.GetOriginalUrlRequest{ShortUrl: "nonexistent"}
	repo.On("GetUrl", "nonexistent").Return("", errors.New("short URL not found")) // Мокируем ошибку

	_, err := service.GetOriginalUrl(context.Background(), req)

	if err == nil || status.Code(err) != codes.NotFound {
		t.Errorf("GetOriginalUrl() expected NotFound error, got: %v", err)
	}

	repo.AssertExpectations(t)
}
