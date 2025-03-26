package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/STaninnat/ecom-backend/internal/database"

	_ "github.com/lib/pq"
)

func (cfg *APIConfig) ConnectDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("Warning: database URL is not set")
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Warning: can't connect to database: %v\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}

	dbQueries := database.New(db)
	cfg.DB = dbQueries
	cfg.DBConn = db

	log.Println("Connected to database successfully...")
}
