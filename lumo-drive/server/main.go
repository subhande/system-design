package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

const defaultBodyLimit = 2 // 2MB

func getBodyLimitBytes() int {
	limitMB := defaultBodyLimit
	v := os.Getenv("BODY_LIMIT_MB")
	if v != "" {
		parsed, err := strconv.Atoi(v)
		if err == nil && parsed > 0 {
			limitMB = parsed
		} else {
			fmt.Printf("Invalid BODY_LIMIT_MB value '%s', using default %d MB\n", v, defaultBodyLimit)
		}
	}

	return limitMB * 1024 * 1024
}

func main() {

	// Load environment variables from .env files if present. Values already set
	// in the real environment take precedence over the file contents.
	_ = godotenv.Load(".env", "../.env")

	connectDB()
	defer DB.Close()

	initDB()

	app := fiber.New(fiber.Config{
		BodyLimit: getBodyLimitBytes(),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default 500
			code := fiber.StatusInternalServerError

			// If it's a fiber error, get status code
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Print actual error in console
			fmt.Println("ERROR:", err)

			// Send response
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${error}\n",
	}))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	folders := v1.Group("/folders")
	files := v1.Group("/files")
	sync := v1.Group("/sync")

	// auth apis (public)
	auth.Post("/register", registerHandler)
	auth.Post("/login", loginHandler)

	// folder apis
	folders.Post("/", AuthMiddleware, createFolderHandler)
	folders.Get("/:folder_id/contents", AuthMiddleware, listFolderContentsHandler)
	folders.Get("/:folder_id", AuthMiddleware, getFolderHandler)
	folders.Patch("/:folder_id", AuthMiddleware, renameFolderHandler)
	folders.Post("/:folder_id/move", AuthMiddleware, moveFolderHandler)
	folders.Delete("/:folder_id", AuthMiddleware, deleteFolderHandler)

	// file apis
	files.Post("/", AuthMiddleware, initiateUploadHandler)
	files.Get("/:file_id/chunks/:chunk_index/presign", AuthMiddleware, getPresignedURLHandler)
	files.Patch("/:file_id/status/:status", AuthMiddleware, updateFileStatusHandler)
	files.Patch("/:file_id/chunks/:chunk_index/status/:status", AuthMiddleware, updateFileChunkStatusHandler)
	files.Post("/:file_id/resume", AuthMiddleware, getPendingChunksHandler)
	files.Get("/:file_id", AuthMiddleware, fetchFileMetadataHandler)
	files.Get("/:file_id/download", AuthMiddleware, generateDownloadURLHandler)
	files.Patch("/:file_id/rename", AuthMiddleware, renameFileHandler)
	files.Post("/:file_id/move", AuthMiddleware, moveFileHandler)
	files.Delete("/:file_id", AuthMiddleware, deleteFileHandler)

	// sync api
	sync.Get("/changes/:change_id", AuthMiddleware, syncHandler)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	err := app.Listen(":6000")
	if err != nil {
		log.Fatal(err)
	}

}
