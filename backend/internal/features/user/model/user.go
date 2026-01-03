package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID
	Username        string
	PasswordHash    string
	AIPrompt        *string
	DiscordWebhook  *string
	NotifyThreshold int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UserJobMatch struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	JobID     uuid.UUID
	Score     int
	Analysis  map[string]interface{}
	Notified  bool
	CreatedAt time.Time
}

type Preference struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
