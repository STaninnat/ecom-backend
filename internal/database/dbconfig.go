package database

import (
	"database/sql"
	"log"
	"os"
)

func ConnectDB() *Queries {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("Warning: database URL is not set")
		return nil
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Warning: can't connect to database: %v\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}

	log.Println("Connected to database successfully...")
	return New(db)
}
