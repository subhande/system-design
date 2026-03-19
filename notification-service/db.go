package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

var initScriptPath = "schema.sql"

func runInitScript(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var statement strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		statement.WriteString(line + "\n")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			_, err := DB.Exec(context.Background(), statement.String())
			if err != nil {
				return err
			}
			statement.Reset()
		}
	}

	return scanner.Err()
}

func ConnectDB() {
	loadDotEnv(".env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set; add it to the environment or controller-service/.env")
	}

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to DB:", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping DB:", err)
	}

	log.Println("Connected to PostgreSQL")

	// Run initialization script
	if err := runInitScript(initScriptPath); err != nil {
		log.Fatal("Failed to run initialization script:", err)
	}

	log.Println("Database initialized")
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if err := os.Setenv(key, value); err != nil {
			log.Printf("warning: unable to set %s from %s: %v", key, path, err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: unable to read %s: %v", path, err)
	}
}
