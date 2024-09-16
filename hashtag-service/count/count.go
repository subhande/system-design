package count

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashtag_service/db"
	"github.com/hashtag_service/models"
	"github.com/segmentio/kafka-go"
)

func comsumer() {
	topic := "user-id--post-hashtag"
	groupID := "count-service"
	hashTagMapCount := make(map[string]int)
	client := db.Connect()

	// Create a new consumer
	config := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers: []string{"localhost:9094"},
			Topic:   topic,
			GroupID: groupID,
		},
	)

	defer config.Close()

	// Read the message
	for {
		msg, err := config.ReadMessage(context.Background())
		if err != nil {
			log.Fatal("Error reading message", err)
		}
		data := models.PostWithHashTag{}
		err = json.Unmarshal(msg.Value, &data)
		if err != nil {
			fmt.Println("Error in unmarshalling post")
		}

		if _, ok := hashTagMapCount[data.Hashtag]; ok {
			hashTagMapCount[data.Hashtag]++
		} else {
			hashTagMapCount[data.Hashtag] = 1
		}

		if hashTagMapCount[data.Hashtag] >= 10 {
			db.UpdateCount(client, data.Hashtag, hashTagMapCount[data.Hashtag])
			hashTagMapCount[data.Hashtag] = 0
		}

	}
}

func RunConsumer() {
	for i := 0; i < 209; i++ {
		go comsumer()
	}
	select {}
}
