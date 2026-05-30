package main

import "os"

// bucketName returns the configured S3 bucket. It is read lazily (not at
// package-init time) so values loaded from .env by godotenv.Load in main are
// visible; reading it into a package-level var would capture an empty string
// before the .env is loaded.
func bucketName() string {
	return os.Getenv("BUCKET_NAME")
}
