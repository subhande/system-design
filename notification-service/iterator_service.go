package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"text/template"
)

var legacyTemplatePlaceholderPattern = regexp.MustCompile(`{{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}`)

func IteratorService() {
	queueName := ResolveRabbitMQQueue("bulk-filter")

	rabbitMQ, err := GetRabbitMQClient(queueName)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("iterator service consuming from %s", queueName)

	if err := rabbitMQ.ConsumeBulkFilterMessages(ctx, processBulkFilterJob); err != nil {
		log.Fatal(err)
	}
}

func processBulkFilterJob(job BulkFilterJob) error {
	log.Printf("processing campaign=%d template=%d priority=%s", job.CampaignID, job.TemplateID, job.Priority)

	if err := UpdateCampaignStatus(job.CampaignID, CampaignStatusRunning); err != nil {
		return fmt.Errorf("mark campaign running: %w", err)
	}

	templateRecord, err := GetTemplate(job.TemplateID)
	if err != nil {
		return fmt.Errorf("load template %d: %w", job.TemplateID, err)
	}

	recipients, err := GetCampaignRecipients(job.CampaignID)
	if err != nil {
		return fmt.Errorf("load campaign recipients: %w", err)
	}

	log.Printf("campaign=%d loaded %d recipients", job.CampaignID, len(recipients))

	priorityQueue := ResolveRabbitMQQueue(string(job.Priority))
	priorityRabbitMQ, err := GetRabbitMQClient(priorityQueue)
	if err != nil {
		return fmt.Errorf("connect priority queue %s: %w", priorityQueue, err)
	}

	renderedAny := false

	for _, recipient := range recipients {
		user, err := GetUser(recipient.UserID)
		if err != nil {
			return fmt.Errorf("load recipient user %d: %w", recipient.UserID, err)
		}

		body, err := renderTemplateBody(templateRecord.Content, user)
		if err != nil {
			return fmt.Errorf("render template for user %d: %w", user.ID, err)
		}

		notificationJob := NotificationJob{
			CampaignID:      job.CampaignID,
			RecipientID:     recipient.ID,
			Body:            body,
			Destination:     user.Email,
			DestinationType: "email",
		}

		if err := priorityRabbitMQ.EnqueueMessage(notificationJob); err != nil {
			return fmt.Errorf("enqueue notification for recipient %d: %w", recipient.ID, err)
		}

		if err := UpdateCampaignRecipientStatus(recipient.ID, RecipientStatusProcessing); err != nil {
			return fmt.Errorf("mark recipient processing %d: %w", recipient.ID, err)
		}

		renderedAny = true
	}

	if renderedAny {
		if err := UpdateCampaignStatus(job.CampaignID, CampaignStatusQueued); err != nil {
			return fmt.Errorf("mark campaign queued: %w", err)
		}
		return nil
	}

	if err := UpdateCampaignStatus(job.CampaignID, CampaignStatusCompleted); err != nil {
		return fmt.Errorf("mark campaign completed: %w", err)
	}

	return nil
}

func renderTemplateBody(content string, user *User) (string, error) {
	// Hello {{name}}, welcome to our service! Email: {{email}}.
	// Fill name and email from user record
	normalizedContent := legacyTemplatePlaceholderPattern.ReplaceAllString(content, "{{.$1}}")

	tmpl, err := template.New("notification").Parse(normalizedContent)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := map[string]string{
		"name":  user.Name,
		"email": user.Email,
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return sb.String(), nil

}
