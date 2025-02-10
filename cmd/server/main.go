package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Dor1ma/url-shortener/internal/config"
	"github.com/Dor1ma/url-shortener/internal/shortener/service"
	"github.com/Dor1ma/url-shortener/internal/shortener/storage"
	pb "github.com/Dor1ma/url-shortener/pkg/grpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	var repo storage.Repository
	var err error

	if cfg.StorageType == "postgres" {
		err = waitForDatabase(cfg.DBConnStr)
		if err != nil {
			log.Fatalf("Error waiting for database: %v", err)
		}

		repo, err = storage.NewPostgresRepository(cfg.DBConnStr)
		if err != nil {
			log.Fatalf("failed to create PostgreSQL repository: %v", err)
		}
		defer repo.Close()
	} else if cfg.StorageType == "in_memory" {
		repo = storage.NewInMemoryRepository()
	} else {
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	service := shortener.NewService(repo)

	address := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUrlShortenerServer(grpcServer, service)

	log.Printf("gRPC server is running on port %s", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	waitForShutdownSignal(grpcServer)
}

func waitForShutdownSignal(grpcServer *grpc.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("Shutting down server gracefully...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func waitForDatabase(connStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for database to become available: %v", ctx.Err())
		default:
			db, err := sql.Open("postgres", connStr)
			if err == nil {
				err = db.Ping()
				if err == nil {
					db.Close()
					return nil
				}
			}
			log.Printf("waiting for database to become available - %v", err)
			time.Sleep(1 * time.Second)
		}
	}
}
