package hashtag_extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashtag_service/models"
	"github.com/segmentio/kafka-go"
)

func extractHashtags(data models.PostWithHashTags) {

	var hashTagData []models.PostWithHashTag

	for _, hashtag := range data.Hashtags {
		hashTagData = append(hashTagData, models.PostWithHashTag{
			ID:      data.ID,
			UserID:  data.UserID,
			Hashtag: hashtag,
			URL:     data.URL,
		})
	}

	topic := "user-id--post-hashtag"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9094"},
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})

	defer writer.Close()

	var kafkaMessages []kafka.Message

	for _, hashTag := range hashTagData {
		hashTagJSON, err := json.Marshal(hashTag)
		if err != nil {
			fmt.Println("Error in marshalling post")
		}
		kafkaMessages = append(kafkaMessages, kafka.Message{
			Key:   []byte(hashTag.Hashtag),
			Value: hashTagJSON,
		})
	}

	err := writer.WriteMessages(context.Background(), kafkaMessages...)

	if err != nil {
		fmt.Println("Error in writing message to kafka")
	}

}

func comsumer() {
	topic := "user-id--post-hashtags"
	groupID := "hashtag-extractor"

	// Create a new consumer
	config := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers: []string{"localhost:9094"},
			Topic:   topic,
			// Partition: partition_id,
			GroupID: groupID,
		},
	)
	count := 0
	defer config.Close()

	// Read the message
	for {
		msg, err := config.ReadMessage(context.Background())
		if err != nil {
			log.Fatal("Error reading message", err)
		}
		data := models.PostWithHashTags{}
		err = json.Unmarshal(msg.Value, &data)
		if err != nil {
			fmt.Println("Error in unmarshalling post")
		}
		go extractHashtags(data)

		count++
		fmt.Printf("GroupID: %v, Message count: %d\n", groupID, count)

	}
}

func RunConsumer() {
	for i := 0; i < 10; i++ {
		go comsumer()
	}
	select {}
}
