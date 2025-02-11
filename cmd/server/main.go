package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Dor1ma/url-shortener/api/gen/go"
	"github.com/Dor1ma/url-shortener/internal/config"
	"github.com/Dor1ma/url-shortener/internal/shortener/service"
	"github.com/Dor1ma/url-shortener/internal/shortener/storage"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	go startGRPCServer(cfg.GRPCPort, repo)

	go startHTTPGateway(cfg.GRPCPort, cfg.HTTPPort)

	waitForShutdownSignal()
}

func startGRPCServer(port string, repo storage.Repository) {
	address := ":" + port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUrlShortenerServer(grpcServer, shortener.NewService(repo))

	log.Printf("gRPC server is running on port %s", port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startHTTPGateway(grpcPort, httpPort string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterUrlShortenerHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		log.Fatalf("failed to start HTTP gateway: %v", err)
	}

	log.Printf("HTTP server is running on port %s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func waitForShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server gracefully...")
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
