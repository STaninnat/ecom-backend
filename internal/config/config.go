package config

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type APIConfig struct {
	Port                string
	DB                  *database.Queries
	DBConn              *sql.DB
	RedisClient         redis.Cmdable
	JWTSecret           string
	RefreshSecret       string
	Issuer              string
	Audience            string
	CredsPath           string
	S3Bucket            string
	S3Region            string
	S3Client            *s3.Client
	StripeSecretKey     string
	StripeWebhookSecret string
	MongoClient         *mongo.Client
	MongoDB             *mongo.Database
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

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("Warning: S3 bucket environment variable is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		log.Fatal("Warning: S3 region environment variable is not set")
	}

	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	if s3Region == "" {
		log.Fatal("Warning: stripe secret key environment variable is not set")
	}

	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if s3Region == "" {
		log.Fatal("Warning: stripe webhook secret environment variable is not set")
	}

	redisClient := InitRedis()

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(s3Region),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(awsCfg)

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("Warning: Mongo URI environment variable is not set")
	}
	mongoClient, mongoDB := ConnectMongoDB(mongoURI)

	return &APIConfig{
		Port:                port,
		RedisClient:         redisClient,
		JWTSecret:           jwtSecret,
		RefreshSecret:       refreshSecret,
		Issuer:              issuerName,
		Audience:            audienceName,
		CredsPath:           credsPath,
		S3Bucket:            s3Bucket,
		S3Region:            s3Region,
		S3Client:            client,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		MongoClient:         mongoClient,
		MongoDB:             mongoDB,
	}
}
