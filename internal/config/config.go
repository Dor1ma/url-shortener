package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DBConnStr   string
	HTTPPort    string
	GRPCPort    string
	StorageType string
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: environment variable %s is not set", key)
	}
	return value
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	dbHost := getEnv("POSTGRES_HOST")
	dbPort := getEnv("POSTGRES_PORT")
	dbUser := getEnv("POSTGRES_USER")
	dbPassword := getEnv("POSTGRES_PASSWORD")
	dbName := getEnv("POSTGRES_DB")

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
		log.Println("Info: GRPC_PORT is not set, using default 50051")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
		log.Println("Info: HTTP_PORT is not set, using default 8080")
	}

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "in_memory"
		log.Println("Info: STORAGE_TYPE is not set, using default 'in_memory'")
	}

	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	return &Config{
		DBConnStr:   dbConnStr,
		GRPCPort:    grpcPort,
		HTTPPort:    httpPort,
		StorageType: storageType,
	}
}
