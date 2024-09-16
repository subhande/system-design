package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashtag_service/count"
	"github.com/hashtag_service/db"
	"github.com/hashtag_service/hashtag_extractor"
	"github.com/hashtag_service/post_generator"
)

// post service -> kafka (p1, h1, h2, h3) -> Hashtag extraction consumer -> kafka (p1, h1) -> Counting server -> DB
// hashtag service -> DB

// Kafka 1
// Kafka 2
// DB

// Post Generator
// Hashtag Extraction
// Counting Server
// Hashtag Service

func main() {
	// Get command line arguments
	args := os.Args[1:]
	fmt.Printf("Arguments: %v\n", args)
	if len(args) == 0 {
		fmt.Println("Please provide a service to run")
		return
	} else if args[0] == "producer" {
		go post_generator.AddPostIntoKafka()
	} else if args[0] == "extractor" {
		go hashtag_extractor.RunConsumer()
	} else if args[0] == "init-db" {
		db.InitialData()
		return
	} else if args[0] == "count-svc" {
		count.RunConsumer()
	} else {
		fmt.Println("Invalid service")
		return
	}
	select {}
}
