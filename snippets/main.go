package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

const defaultBodyLimitMB = 10

func getBodyLimitBytes() int {
	limitMB := defaultBodyLimitMB
	v := os.Getenv("BODY_LIMIT_MB")
	if v != "" {
		parsed, err := strconv.Atoi(v)
		if err == nil && parsed > 0 {
			limitMB = parsed
		} else {
			fmt.Printf("Invalid BODY_LIMIT_MB value '%s', using default %d MB\n", v, defaultBodyLimitMB)
		}
	}

	return limitMB * 1024 * 1024
}

const DEFAULT_OWNER_ID = 1
const DEFAULT_VISIBILITY = "public"

var BUCKET_NAME string

var producer *KafkaProducer

// Create folder if not exists
func init() {

	log.Info("Initializing application...")
	godotenv.Load()
	BUCKET_NAME = os.Getenv("BUCKET_NAME")
	if BUCKET_NAME == "" {
		log.Fatal("BUCKET_NAME environment variable is not set")
	}
	log.Debug("Creating BUCKET_NAME directory if it does not exist: %s", BUCKET_NAME)
	if _, err := os.Stat(BUCKET_NAME); os.IsNotExist(err) {
		err := os.Mkdir(BUCKET_NAME, 0755)
		if err != nil {
			log.Fatal("Failed to create bucket directory:", err)
		}
		log.Info("Bucket directory created successfully")
	}
	connectDB()
	initDB()
	log.Debug(fmt.Sprintf("Initializing Kafka producer with brokers: %s and topic: %s", os.Getenv("KAFKA_BROKERS"), os.Getenv("KAFKA_ANALYTICS_TOPIC")))
	producer = NewKafkaProducer(os.Getenv("KAFKA_BROKERS"), os.Getenv("KAFKA_ANALYTICS_TOPIC"))
}

func cleanup() {
	log.Info("Cleaning up resources...")
	if producer != nil {
		producer.Close()
		log.Info("Kafka producer closed successfully")
	}
	if DB != nil {
		DB.Close()
		log.Info("Database connection pool closed successfully")
	}
}

func storeSnippet(id string, snippetPath string, fileName string, content string) (bool, error) {
	// This is mock S3 storage logic. In a real implementation, you would use the AWS SDK to upload the content to an S3 bucket.
	log.Info("Storing snippet with ID %s at path %s", id, snippetPath)

	// Create directory with snippetPath if it does not exist
	if _, err := os.Stat(snippetPath); os.IsNotExist(err) {
		err := os.MkdirAll(snippetPath, 0755)
		if err != nil {
			log.Error("Failed to create directory for snippet with ID %s: %v", id, err)
			return false, err
		}
		log.Info("Directory created successfully for snippet with ID %s at path %s", id, snippetPath)
	}

	// Save to the local filesystem as text file for demonstration purposes
	err := os.WriteFile(snippetPath+"/"+fileName, []byte(content), 0644)
	if err != nil {
		log.Error("Failed to store snippet with ID %s: %v", id, err)
		return false, err
	}

	log.Info("Snippet with ID %s stored successfully at %s", id, snippetPath+"/"+fileName)
	return true, nil
}

func retriveSnippet(id string, snippetPath string, fileName string) (string, error) {
	// This is mock S3 retrieval logic. In a real implementation, you would use the AWS SDK to download the content from an S3 bucket.
	log.Info("Retrieving snippet with ID %s from path %s", id, snippetPath+"/"+fileName)

	// Read from the local filesystem for demonstration purposes
	content, err := os.ReadFile(snippetPath + "/" + fileName)
	if err != nil {
		log.Error("Failed to retrieve snippet with ID %s: %v", id, err)
		return "", err
	}

	log.Info("Snippet with ID %s retrieved successfully from %s", id, snippetPath+"/"+fileName)
	return string(content), nil
}

func createSnippetHandler(c *fiber.Ctx) error {

	newSnippet := new(CreateSnippetRequest)
	if err := c.BodyParser(newSnippet); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	// UUID V4
	newSnippet.UUID = uuid.NewString()

	if newSnippet.Name == "" {
		// Generate a random name
		newSnippet.Name = fmt.Sprintf("snippet-%d", time.Now().Unix())
	}

	if newSnippet.Visibility == "" {
		newSnippet.Visibility = DEFAULT_VISIBILITY
	}

	if newSnippet.OwnerID == 0 {
		newSnippet.OwnerID = DEFAULT_OWNER_ID
	}

	var contentPath string
	contentPath = fmt.Sprintf("%s/%d", BUCKET_NAME, newSnippet.OwnerID)

	var fileName string
	fileName = newSnippet.UUID + ".txt"

	_, err := storeSnippet(newSnippet.UUID, contentPath, fileName, newSnippet.Content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to store snippet content")
	}
	var expiredAt *time.Time
	if newSnippet.Expiry != 0 {
		t := time.Now().Add(time.Duration(newSnippet.Expiry) * 24 * time.Hour)
		expiredAt = &t
	}

	// Save to Database
	_, err = DB.Exec(context.Background(), `
		INSERT INTO snippets (uuid, name, visibility, owner_id, expiredAt)	
		VALUES ($1, $2, $3, $4, $5)
	`, newSnippet.UUID, newSnippet.Name, newSnippet.Visibility, newSnippet.OwnerID, expiredAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save snippet metadata to database")
	}

	log.Info("Snippet with ID %s created successfully in database", newSnippet.UUID)

	contentSizeInKB := len(newSnippet.Content) / 1024
	log.Info("Retrieved snippet content size: %d KB", contentSizeInKB)

	event := AnalyticsEvent{
		EventType:       EventTypeSnippetCreated,
		SnippetID:       newSnippet.UUID,
		OwnerID:         newSnippet.OwnerID,
		Visibility:      newSnippet.Visibility,
		RequestPath:     c.Path(),
		UserAgent:       c.Get("User-Agent"),
		IP:              c.IP(),
		EventTime:       time.Now(),
		ContentSizeInKB: contentSizeInKB,
	}

	err = producer.Publish(context.Background(), event)
	if err != nil {
		log.Error("Failed to publish analytics event for snippet creation: %v", err)
	}

	return c.JSON(fiber.Map{
		"message":    "Create a new paste bin snippet",
		"snippet_id": newSnippet.UUID,
	})
}

func getSnippetHandler(c *fiber.Ctx) error {
	// Path Params | :owner_id/:snippet_id
	owner_id := c.Params("owner_id", "0")
	snippet_id := c.Params("snippet_id")

	// DB Query Fetch the snippet using snippet_id

	query := `SELECT * FROM snippets WHERE uuid = $1`
	row := DB.QueryRow(context.Background(), query, snippet_id)

	var snippet Snippet
	err := row.Scan(&snippet.UUID, &snippet.Name, &snippet.CreatedAt, &snippet.Visibility, &snippet.OwnerID, &snippet.ExpiredAt)
	if err != nil {
		log.Error("Failed to fetch snippet with ID %s: %v", snippet_id, err)
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Snippet with ID %s not found", snippet_id))
	}

	if snippet.ExpiredAt != nil && !snippet.ExpiredAt.IsZero() && time.Now().After(*snippet.ExpiredAt) {
		return fiber.NewError(fiber.StatusGone, "Snippet has expired")
	}

	if strconv.Itoa(snippet.OwnerID) != owner_id && snippet.Visibility == "private" {
		return fiber.NewError(fiber.StatusForbidden, "You do not have access to this snippet")
	}

	var snippetPath string
	snippetPath = fmt.Sprintf("%s/%d", BUCKET_NAME, snippet.OwnerID)

	var fileName string
	fileName = snippet_id + ".txt"

	content, err := retriveSnippet(snippet.UUID, snippetPath, fileName)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve snippet content")
	}

	contentSizeInKB := len(content) / 1024
	log.Info("Retrieved snippet content size: %d KB", contentSizeInKB)

	event := AnalyticsEvent{
		EventType:       EventTypeSnippetViewed,
		SnippetID:       snippet.UUID,
		OwnerID:         snippet.OwnerID,
		Visibility:      snippet.Visibility,
		ContentSizeInKB: contentSizeInKB,
		RequestPath:     c.Path(),
		UserAgent:       c.Get("User-Agent"),
		IP:              c.IP(),
		EventTime:       time.Now(),
	}

	err = producer.Publish(context.Background(), event)
	if err != nil {
		log.Error("Failed to publish analytics event for snippet creation: %v", err)
	}

	return c.JSON(fiber.Map{
		"uuid":       snippet.UUID,
		"name":       snippet.Name,
		"created_at": snippet.CreatedAt,
		"visibility": snippet.Visibility,
		"owner_id":   snippet.OwnerID,
		"content":    content,
		"expired_at": snippet.ExpiredAt,
	})
}

func main() {
	defer cleanup()
	// Start multiple Kafka consumers to handle analytics events concurrently
	for i := 0; i < 3; i++ {
		// Start Kafka consumer in a separate goroutine
		go kafkaConsumeAnalyticsEvents(context.Background())
	}
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

	snippetsRoute := v1.Group("/snippets")
	snippetsRoute.Get(":snippet_id/:owner_id", getSnippetHandler)
	snippetsRoute.Get(":snippet_id", getSnippetHandler)
	snippetsRoute.Post("/", createSnippetHandler)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to the Snippets API! Use /api/v1/snippets to create and retrieve snippets.",
			"endpoints": []string{
				"POST /api/v1/snippets - Create a new snippet",
				"GET /api/v1/snippets/:owner_id/:snippet_id - Retrieve a snippet",
			},
			"example": map[string]string{
				"create_snippet": `curl -X POST http://localhost:6000/api/v1/snippets -H "Content-Type: application/json" -d '{"name":"My Snippet","content":"This is the content of the snippet.","visibility":"public","owner_id":1}'`,
				"get_snippet":    `curl http://localhost:6000/api/v1/snippets/1/<snippet_uuid>`,
			},
			"note":         "Replace <snippet_uuid> with the actual UUID returned when you create a snippet.",
			"health_check": "Use /health endpoint to check if the service is running.",
		})
	})

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
