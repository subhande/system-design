package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func connectDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	var err error
	DB, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	log.Println("Connected to database successfully")
}

func initDB() {
	// uid (uuid), name (varchar), createdAt (timestamp), visibility (varchar), owner_id (int)
	_, err := DB.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS snippets (
        uuid UUID PRIMARY KEY,
        name VARCHAR(120) NOT NULL,
        createdAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        visibility VARCHAR(20) NOT NULL,
        owner_id int NOT NULL
    );
    `)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	log.Println("Database initialized successfully")
}
