package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool

func connectDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	var err error
	DB, err = pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	log.Println("Connected to database successfully")
}

func cleanUpDB() {
	// Read cleanup.sql file
	cleanupSQLBytes, err := os.ReadFile("cleanup.sql")
	if err != nil {
		log.Fatal("Failed to read cleanup.sql file:", err)
	}
	cleanupSQL := string(cleanupSQLBytes)

	_, err = DB.Exec(context.Background(), cleanupSQL)

	if err != nil {
		log.Fatal("Failed to clean up database:", err)
	}
	log.Println("Database cleaned up successfully")
}

func initDB() {

	// cleanUpDB()

	// Execute init.sql file
	initSQLBytes, err := os.ReadFile("init.sql")
	if err != nil {
		log.Fatal("Failed to read init.sql file:", err)
	}
	initSQL := string(initSQLBytes)

	_, err = DB.Exec(context.Background(), initSQL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	log.Println("Database initialized successfully")

}
