package config

import (
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/redis/go-redis/v9"
)

type APIConfig struct {
	DB          *database.Queries
	RedisClient *redis.Client
	JWTSecret   string
	Issuer      string
	Audience    string
}
