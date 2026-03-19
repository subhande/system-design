package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func WorkerService() {
	priority := "p1"
	if len(os.Args) > 2 && strings.TrimSpace(os.Args[2]) != "" {
		priority = os.Args[2]
	}

	queueName := ResolveRabbitMQQueue(priority)

	rabbitMQ, err := GetRabbitMQClient(queueName)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker service consuming from %s", queueName)

	if err := rabbitMQ.ConsumeMessages(ctx, processNotificationJob); err != nil {
		log.Fatal(err)
	}
}

func processNotificationJob(job NotificationJob) error {
	if job.CampaignID == 0 || job.RecipientID == 0 {
		return fmt.Errorf("notification job is missing campaign_id or recipient_id")
	}

	if err := UpdateCampaignStatus(job.CampaignID, CampaignStatusRunning); err != nil {
		return fmt.Errorf("mark campaign running: %w", err)
	}

	providerMessageID, err := deliverNotification(job)
	if err != nil {
		if updateErr := MarkCampaignRecipientFailed(job.RecipientID, err.Error()); updateErr != nil {
			return fmt.Errorf("delivery failed: %v; mark recipient failed: %w", err, updateErr)
		}
	} else {
		if err := MarkCampaignRecipientSent(job.RecipientID, providerMessageID); err != nil {
			return fmt.Errorf("mark recipient sent: %w", err)
		}
	}

	if err := SyncCampaignDeliveryStats(job.CampaignID); err != nil {
		return fmt.Errorf("sync campaign delivery stats: %w", err)
	}

	return nil
}

func deliverNotification(job NotificationJob) (string, error) {
	destination := strings.TrimSpace(job.Destination)
	destinationType := strings.ToLower(strings.TrimSpace(job.DestinationType))

	if destination == "" {
		return "", fmt.Errorf("missing destination")
	}

	switch destinationType {
	case "email", "sms", "push":
	default:
		return "", fmt.Errorf("unsupported destination type %q", job.DestinationType)
	}

	// Simulate provider work so the worker behaves like a real consumer.
	time.Sleep(150 * time.Millisecond)

	// Keep the example deterministic: every fifth recipient fails delivery.
	if job.RecipientID%5 == 0 {
		return "", fmt.Errorf("simulated %s provider failure", destinationType)
	}

	providerMessageID := fmt.Sprintf("%s-%d-%d", destinationType, job.CampaignID, job.RecipientID)
	log.Printf(
		"sent %s notification for campaign=%d recipient=%d destination=%s message_id=%s",
		destinationType,
		job.CampaignID,
		job.RecipientID,
		destination,
		providerMessageID,
	)

	return providerMessageID, nil
}
