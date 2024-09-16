package post_generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashtag_service/models"
	"github.com/segmentio/kafka-go"
	// Add this import statement
)

func getHashtags() string {

	// Read hashtag.json file
	file, err := os.Open("data/hashtags.json")
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer file.Close()

	// Read the contents of the file
	byteValue, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}

	// Optionally, parse the JSON data
	var hashtags map[string][]string
	err = json.Unmarshal(byteValue, &hashtags)
	if err != nil {
		log.Fatalf("failed to unmarshal JSON: %s", err)
	}
	// fmt.Println(hashtags)

	var hashTags []string = hashtags["hashtags"]

	// get a random hashtag
	return hashTags[rand.Intn(len(hashTags))]
}

func getUserId() string {
	var NO_OF_USERS int = 10
	userId := rand.Intn(NO_OF_USERS) + 1
	return strconv.Itoa(userId)
}

func postGenerator() ([]byte, models.PostWithHashTags) {

	// 1-7 hashtags
	hashTagLength := rand.Intn(7) + 1

	var hashtags []string

	for i := 0; i < hashTagLength; i++ {
		hashtags = append(hashtags, getHashtags())
	}

	post := models.PostWithHashTags{
		ID:       uuid.New().String(),
		UserID:   getUserId(),
		Hashtags: hashtags,
		URL:      "https://www.instagram.com/p/" + uuid.New().String(),
	}

	post_json, err := json.Marshal(post)
	if err != nil {
		fmt.Println("Error in marshalling post")
	}
	return post_json, post
}

func AddPostIntoKafka() {

	topic := "user-id--post-hashtags"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9094"},
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})

	defer writer.Close()

	for {
		kafkaMessages := make([]kafka.Message, 100)

		for i := 0; i < 100; i++ {
			post_json, post := postGenerator()
			fmt.Printf("Inserting post %d: %v\n", i, post)
			kafkaMessages[i] = kafka.Message{
				Key:   []byte(post.UserID),
				Value: post_json,
			}

		}
		err := writer.WriteMessages(context.Background(),
			kafkaMessages...,
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
	}
}
