package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultRabbitMQQueue = "notification_jobs"

const (
	RabbitMQQueueBulkFilterEnv = "RABBITMQ_BULKFILTER_QUEUE"
	RabbitMQQueueIteratorP1Env = "RABBITMQ_ITERATOR_QUEUE_P1"
	RabbitMQQueueIteratorP2Env = "RABBITMQ_ITERATOR_QUEUE_P2"
	RabbitMQQueueIteratorP3Env = "RABBITMQ_ITERATOR_QUEUE_P3"
)

type RabbitMQClient struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

var (
	rabbitMQClientsMu sync.Mutex
	rabbitMQClients   = make(map[string]*RabbitMQClient)
)

type RabbitMQQueues struct {
	BulkFilter string
	IteratorP1 string
	IteratorP2 string
	IteratorP3 string
}

func deadLetterQueueName(queueName string) string {
	return fmt.Sprintf("%s.dlq", queueName)
}

func LoadRabbitMQQueues() RabbitMQQueues {
	return RabbitMQQueues{
		BulkFilter: queueNameFromEnv(RabbitMQQueueBulkFilterEnv, "bulk-filter-queue"),
		IteratorP1: queueNameFromEnv(RabbitMQQueueIteratorP1Env, "iterator-queue-p1"),
		IteratorP2: queueNameFromEnv(RabbitMQQueueIteratorP2Env, "iterator-queue-p2"),
		IteratorP3: queueNameFromEnv(RabbitMQQueueIteratorP3Env, "iterator-queue-p3"),
	}
}

func queueNameFromEnv(envKey string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(envKey))
	if value != "" {
		return value
	}
	return fallback
}

func ConnectRabbitMQ(queueName string) (*RabbitMQClient, error) {
	loadDotEnv(".env")

	dsn := os.Getenv("RABBITMQ_URL")
	if dsn == "" {
		dsn = "amqp://admin:1234@localhost:5672/"
	}

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create rabbitmq channel: %w", err)
	}

	if queueName == "" {
		queueName = defaultRabbitMQQueue
	}

	if err := declareDurableQueue(channel, queueName); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Process one message at a time so manual ack reflects completed work.
	if err := channel.Qos(1, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("configure rabbitmq qos: %w", err)
	}

	return &RabbitMQClient{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func GetRabbitMQClient(queueName string) (*RabbitMQClient, error) {
	queueName = ResolveRabbitMQQueue(queueName)

	rabbitMQClientsMu.Lock()
	defer rabbitMQClientsMu.Unlock()

	if client, ok := rabbitMQClients[queueName]; ok {
		return client, nil
	}

	client, err := ConnectRabbitMQ(queueName)
	if err != nil {
		return nil, err
	}

	rabbitMQClients[queueName] = client
	return client, nil
}

func CloseAllRabbitMQClients() error {
	rabbitMQClientsMu.Lock()
	defer rabbitMQClientsMu.Unlock()

	var closeErr error
	for queueName, client := range rabbitMQClients {
		if err := client.Close(); err != nil && closeErr == nil {
			closeErr = fmt.Errorf("close rabbitmq client for %s: %w", queueName, err)
		}
		delete(rabbitMQClients, queueName)
	}

	return closeErr
}

func ResolveRabbitMQQueue(name string) string {
	queues := LoadRabbitMQQueues()

	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "default":
		return defaultRabbitMQQueue
	case "bulk-filter", "bulkfilter":
		return queues.BulkFilter
	case "iterator-p1", "p1":
		return queues.IteratorP1
	case "iterator-p2", "p2":
		return queues.IteratorP2
	case "iterator-p3", "p3":
		return queues.IteratorP3
	default:
		return name
	}
}

func declareDurableQueue(channel *amqp.Channel, queueName string) error {
	_, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare durable queue %q: %w", queueName, err)
	}

	return nil
}

func (r *RabbitMQClient) EnqueueMessage(job any) error {
	switch job.(type) {
	case BulkFilterJob, NotificationJob, DeadLetterMessage:
	default:
		return fmt.Errorf("unsupported job type %T", job)
	}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal notification job: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish persistent message: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) publishRaw(queueName string, body []byte) error {
	if err := declareDurableQueue(r.channel, queueName); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish persistent message: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) publishDeadLetter(jobType string, originalBody []byte, handlerErr error) error {
	deadLetter := DeadLetterMessage{
		OriginalQueue: r.queueName,
		JobType:       jobType,
		ErrorMessage:  handlerErr.Error(),
		FailedAt:      time.Now(),
		Payload:       json.RawMessage(originalBody),
	}

	body, err := json.Marshal(deadLetter)
	if err != nil {
		return fmt.Errorf("marshal dead letter message: %w", err)
	}

	dlqName := deadLetterQueueName(r.queueName)
	if err := r.publishRaw(dlqName, body); err != nil {
		return fmt.Errorf("publish dead letter message to %s: %w", dlqName, err)
	}

	log.Printf("published failed %s message from %s to %s", jobType, r.queueName, dlqName)
	return nil
}

func (r *RabbitMQClient) ConsumeBulkFilterMessages(ctx context.Context, handler func(BulkFilterJob) error) error {
	deliveries, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register rabbitmq consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq deliveries channel closed")
			}

			var job BulkFilterJob
			if err := json.Unmarshal(delivery.Body, &job); err != nil {
				log.Printf("bulk filter job decode failed on queue=%s: %v", r.queueName, err)
				if dlqErr := r.publishDeadLetter("bulk_filter", delivery.Body, fmt.Errorf("decode message body: %w", err)); dlqErr != nil {
					_ = delivery.Nack(false, true)
					return dlqErr
				}
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack dead-lettered message: %w", ackErr)
				}
				continue
			}

			log.Printf("received bulk filter job campaign_id=%d template_id=%d priority=%s", job.CampaignID, job.TemplateID, job.Priority)

			if err := handler(job); err != nil {
				log.Printf("bulk filter job failed campaign_id=%d: %v", job.CampaignID, err)
				if dlqErr := r.publishDeadLetter("bulk_filter", delivery.Body, err); dlqErr != nil {
					if nackErr := delivery.Nack(false, true); nackErr != nil {
						return fmt.Errorf("handler failed: %v; dead-letter failed: %v; nack failed: %w", err, dlqErr, nackErr)
					}
					continue
				}
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack dead-lettered message: %w", ackErr)
				}
				continue
			}

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack message: %w", err)
			}
		}
	}
}

func (r *RabbitMQClient) ConsumeMessages(ctx context.Context, handler func(NotificationJob) error) error {
	deliveries, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("register rabbitmq consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq deliveries channel closed")
			}

			var job NotificationJob
			if err := json.Unmarshal(delivery.Body, &job); err != nil {
				log.Printf("notification job decode failed on queue=%s: %v", r.queueName, err)
				if dlqErr := r.publishDeadLetter("notification", delivery.Body, fmt.Errorf("decode message body: %w", err)); dlqErr != nil {
					_ = delivery.Nack(false, true)
					return dlqErr
				}
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack dead-lettered message: %w", ackErr)
				}
				continue
			}

			log.Printf("received notification job campaign_id=%d recipient_id=%d destination=%s", job.CampaignID, job.RecipientID, job.Destination)

			if err := handler(job); err != nil {
				log.Printf("notification job failed campaign_id=%d recipient_id=%d: %v", job.CampaignID, job.RecipientID, err)
				if dlqErr := r.publishDeadLetter("notification", delivery.Body, err); dlqErr != nil {
					if nackErr := delivery.Nack(false, true); nackErr != nil {
						return fmt.Errorf("handler failed: %v; dead-letter failed: %v; nack failed: %w", err, dlqErr, nackErr)
					}
					continue
				}
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack dead-lettered message: %w", ackErr)
				}
				continue
			}

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack message: %w", err)
			}
		}
	}
}

func (r *RabbitMQClient) Close() error {
	var closeErr error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			closeErr = err
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}
