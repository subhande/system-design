package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaProducer(kafka_brokers string, kafka_topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(kafka_brokers),
			Topic:        kafka_topic,
			Balancer:     &kafka.LeastBytes{},
			BatchSize:    100, // batch messages
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (p *KafkaProducer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Println("error closing writer:", err)
	}
}

func NewKafkaConsumer(kafka_brokers string, kafka_topic string, kafka_analytics_group string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{kafka_brokers},
			GroupID:  kafka_analytics_group,
			Topic:    kafka_topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

func (c *KafkaConsumer) Close() {
	if err := c.reader.Close(); err != nil {
		log.Println("error closing reader:", err)
	}
}

func (c *KafkaConsumer) Consume(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (p *KafkaProducer) Publish(ctx context.Context, event AnalyticsEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(event.SnippetID),
			Value: value,
		},
	)
}

type ElasticSearchConfig struct {
	Addresses []string
	Username  string
	Password  string
}

type ElasticSearchClient struct {
	Client *elasticsearch.Client
}

func NewElasticSearchClient(config ElasticSearchConfig) (*ElasticSearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ElasticSearchClient{Client: es}, nil
}

func (es *ElasticSearchClient) IndexAnalyticsEvent(ctx context.Context, index string, event AnalyticsEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	res, err := es.Client.Index(
		index,
		bytes.NewReader(data),
		es.Client.Index.WithContext(ctx),
		es.Client.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func kafkaConsumeAnalyticsEvents(ctx context.Context) {

	log.Println("Starting Kafka consumer for analytics events...")

	kafka_brokers := os.Getenv("KAFKA_BROKERS")
	kafka_topic := os.Getenv("KAFKA_ANALYTICS_TOPIC")
	kafka_analytics_group := os.Getenv("KAFKA_ANALYTICS_GROUP")

	elasticSearchURL := os.Getenv("ELASTICSEARCH_URL")
	elasticSearchUsername := os.Getenv("ELASTICSEARCH_USERNAME")
	elasticSearchPassword := os.Getenv("ELASTICSEARCH_PASSWORD")

	esConfig := ElasticSearchConfig{
		Addresses: []string{elasticSearchURL},
		Username:  elasticSearchUsername,
		Password:  elasticSearchPassword,
	}

	esClient, err := NewElasticSearchClient(esConfig)
	if err != nil {
		log.Fatalf("Failed to create ElasticSearch client: %v", err)
	}

	consumer := NewKafkaConsumer(kafka_brokers, kafka_topic, kafka_analytics_group)
	defer consumer.Close()

	for {
		msg, err := consumer.Consume(ctx)
		if err != nil {
			log.Printf("Error consuming message: %v", err)
			continue
		}

		var event AnalyticsEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		es_index := kafka_topic + "-" + time.Now().Format("2006-01-02")

		if err := esClient.IndexAnalyticsEvent(ctx, es_index, event); err != nil {
			log.Printf("Error indexing analytics event: %v", err)
		} else {
			log.Printf("Successfully indexed analytics event for snippet ID: %s", event.SnippetID)
		}
	}

}
