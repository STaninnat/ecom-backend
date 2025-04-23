package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/redis/go-redis/v9"
)

type APIConfig struct {
	Port          string
	DB            *database.Queries
	DBConn        *sql.DB
	RedisClient   redis.Cmdable
	JWTSecret     string
	RefreshSecret string
	Issuer        string
	Audience      string
	CredsPath     string
}

func LoadConfig() *APIConfig {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Warning: Port environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Warning: JWT Secret environment variable is not set")
	}

	refreshSecret := os.Getenv("REFRESH_SECRET")
	if jwtSecret == "" {
		log.Fatal("Warning: Refresh Secret environment variable is not set")
	}

	issuerName := os.Getenv("ISSUER")
	if issuerName == "" {
		log.Fatal("Warning: Issuer environment variable is not set")
	}

	audienceName := os.Getenv("AUDIENCE")
	if audienceName == "" {
		log.Fatal("Warning: Audience environment variable is not set")
	}

	credsPath := os.Getenv("GOOGLE_CREDENTIALS_PATH")
	if credsPath == "" {
		log.Fatal("Warning: Google credentials path environment variable is not set")
	}

	redisClient := InitRedis()

	return &APIConfig{
		Port:          port,
		RedisClient:   redisClient,
		JWTSecret:     jwtSecret,
		RefreshSecret: refreshSecret,
		Issuer:        issuerName,
		Audience:      audienceName,
		CredsPath:     credsPath,
	}
}
