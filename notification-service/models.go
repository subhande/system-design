package main

import (
	"encoding/json"
	"time"
)

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusQueued    CampaignStatus = "queued"
	CampaignStatusRunning   CampaignStatus = "running"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type RecipientStatus string

const (
	RecipientStatusPending    RecipientStatus = "pending"
	RecipientStatusProcessing RecipientStatus = "processing"
	RecipientStatusSent       RecipientStatus = "sent"
	RecipientStatusFailed     RecipientStatus = "failed"
	RecipientStatusSkipped    RecipientStatus = "skipped"
)

type CampaignRecipientType string

const (
	CampaignRecipientTypeAllUsers      CampaignRecipientType = "all_users"
	CampaignRecipientTypeSpecificUsers CampaignRecipientType = "specific_users"
)

type CampaignPriority string

const (
	CampaignPriorityP1 CampaignPriority = "p1"
	CampaignPriorityP2 CampaignPriority = "p2"
	CampaignPriorityP3 CampaignPriority = "p3"
)

type Template struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Campaign struct {
	ID              int64                 `json:"id"`
	UserID          int64                 `json:"user_id"`
	TemplateID      int64                 `json:"template_id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Status          CampaignStatus        `json:"status"`
	RecipientType   CampaignRecipientType `json:"recipient_type"`
	Priority        CampaignPriority      `json:"priority"`
	TotalRecipients int64                 `json:"total_recipients"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
	ScheduledAt     *time.Time            `json:"scheduled_at,omitempty"`
	CompletedAt     *time.Time            `json:"completed_at,omitempty"`
	SentCount       int64                 `json:"sent_count"`
	FailedCount     int64                 `json:"failed_count"`
	SkippedCount    int64                 `json:"skipped_count"`
}

type CampaignRecipient struct {
	ID                int64           `json:"id"`
	CampaignID        int64           `json:"campaign_id"`
	UserID            int64           `json:"user_id"`
	Status            RecipientStatus `json:"status"`
	ProviderMessageID *string         `json:"provider_message_id,omitempty"`
	ErrorMessage      *string         `json:"error_message,omitempty"`
	RetryCount        int32           `json:"retry_count"`
	ScheduledAt       *time.Time      `json:"scheduled_at,omitempty"`
	SentAt            *time.Time      `json:"sent_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BulkFilterJob struct {
	CampaignID    int64                 `json:"campaign_id"`
	TemplateID    int64                 `json:"template_id"`
	UserId        int64                 `json:"user_id"`
	Name          string                `json:"name"`
	RecipientType CampaignRecipientType `json:"recipient_type"`
	Priority      CampaignPriority      `json:"priority"`
}

type NotificationJob struct {
	CampaignID      int64  `json:"campaign_id,omitempty"`
	RecipientID     int64  `json:"recipient_id,omitempty"`
	Body            string `json:"body"`
	Destination     string `json:"destination"`
	DestinationType string `json:"destination_type"`
}

type DeadLetterMessage struct {
	OriginalQueue string          `json:"original_queue"`
	JobType       string          `json:"job_type"`
	ErrorMessage  string          `json:"error_message"`
	FailedAt      time.Time       `json:"failed_at"`
	Payload       json.RawMessage `json:"payload"`
}
