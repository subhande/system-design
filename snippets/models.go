package main

import "time"

type EventType string

const (
	EventTypeSnippetCreated EventType = "snippet_created"
	EventTypeSnippetViewed  EventType = "snippet_viewed"
	EventTypeSnippetDeleted EventType = "snippet_deleted"
)

type AnalyticsEvent struct {
	EventType       EventType
	SnippetID       string
	OwnerID         int
	Visibility      VisibilityType
	ContentSizeInKB int
	RequestPath     string
	UserAgent       string
	IP              string
	EventTime       time.Time
}

type VisibilityType string

const (
	VisibilityPublic  VisibilityType = "public"
	VisibilityPrivate VisibilityType = "private"
)

type CreateSnippetRequest struct {
	UUID       string         `json:"uuid,omitempty"`
	Name       string         `json:"name"`
	Content    string         `json:"content,omitempty"`
	Visibility VisibilityType `json:"visibility,omitempty"` // e.g., "public", "private"
	OwnerID    int            `json:"owner_id,omitempty"`
	Expiry     int            `json:"expiry,omitempty"` // Expiry time in days
}

type Snippet struct {
	UUID       string         `json:"uuid"`
	Name       string         `json:"name"`
	CreatedAt  time.Time      `json:"created_at"`
	ExpiredAt  *time.Time     `json:"expired_at,omitempty"`
	Visibility VisibilityType `json:"visibility"`
	OwnerID    int            `json:"owner_id"`
}
