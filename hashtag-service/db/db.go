package db

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hashtag_service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	return client
}

func Disconnect(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client.Disconnect(ctx)
}

func InitDB() *mongo.Client {
	client := Connect()
	return client
}

func getAllHashtags() []string {

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
	return hashTags
}

func InitialData() {
	client := InitDB()

	defer Disconnect(client)

	collection := client.Database("instaphoto").Collection("top_posts")

	hashtags := getAllHashtags()

	// drop the collection
	collection.Drop(context.Background())

	// create the collection
	collection = client.Database("instaphoto").Collection("top_posts")

	data := []models.HashtagCount{}

	for _, hashtag := range hashtags {
		posts := []models.Post{}
		for i := 0; i < 100; i++ {
			posts = append(posts, models.Post{
				ID:  uuid.New().String(),
				URL: "https://www.instagram.com/p/" + uuid.New().String(),
			})
		}
		data = append(data, models.HashtagCount{
			Hashtag:  hashtag,
			Count:    100,
			TopPosts: posts,
		})
	}

	// convert the data to interface
	var interfaceData []interface{} = make([]interface{}, len(data))
	for i, v := range data {
		d, _ := json.Marshal(v)
		var m interface{}
		json.Unmarshal(d, &m)
		interfaceData[i] = m
	}

	// insert many documents
	_, err := collection.InsertMany(context.Background(), interfaceData)

	if err != nil {
		log.Fatalf("failed to insert documents: %s", err)
	}

}

func UpdateCount(client *mongo.Client, hashtag string, inc_count int) {
	collection := client.Database("instaphoto").Collection("top_posts")

	// filter by hashtag and increase the count by inc_count

	_, err := collection.UpdateOne(context.Background(), bson.M{"hashtag": hashtag}, bson.M{"$inc": bson.M{"count": inc_count}})

	if err != nil {
		log.Fatalf("failed to update count: %s", err)
	}

}
