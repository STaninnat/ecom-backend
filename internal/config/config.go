package config

import (
	"log"
	"os"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type APIConfig struct {
	Port        string
	DB          *database.Queries
	RedisClient *redis.Client
	JWTSecret   string
	Issuer      string
	Audience    string
}

func LoadConfig() *APIConfig {
	err := godotenv.Load(".env.development")
	if err != nil {
		log.Printf("Warning: assuming default configuration, env unreadable: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Warning: Port environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Warning: JWT Secret environment variable is not set")
	}

	issuerName := os.Getenv("ISSUER")
	if issuerName == "" {
		log.Fatal("Warning: Issuer environment variable is not set")
	}

	audienceName := os.Getenv("AUDIENCE")
	if audienceName == "" {
		log.Fatal("Warning: Audience environment variable is not set")
	}

	redisClient := InitRedis()

	return &APIConfig{
		Port:        port,
		RedisClient: redisClient,
		JWTSecret:   jwtSecret,
		Issuer:      issuerName,
		Audience:    audienceName,
	}
}
