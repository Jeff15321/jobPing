package handler

import "github.com/google/uuid"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type CreatePreferenceRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdatePreferenceRequest struct {
	Value string `json:"value"`
}

type PreferenceResponse struct {
	ID    uuid.UUID `json:"id"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type PreferencesResponse struct {
	Preferences []PreferenceResponse `json:"preferences"`
}

type UpdatePromptRequest struct {
	Prompt string `json:"prompt"`
}

type UpdateDiscordRequest struct {
	WebhookURL string `json:"webhook_url"`
}

type UpdateThresholdRequest struct {
	Threshold int `json:"threshold"`
}

type ProfileResponse struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	AIPrompt        *string   `json:"ai_prompt"`
	DiscordWebhook  *string   `json:"discord_webhook"`
	NotifyThreshold int       `json:"notify_threshold"`
}

type UserJobMatchResponse struct {
	ID        uuid.UUID              `json:"id"`
	JobID     uuid.UUID              `json:"job_id"`
	Score     int                    `json:"score"`
	Analysis  map[string]interface{} `json:"analysis"`
	Notified  bool                   `json:"notified"`
	CreatedAt string                 `json:"created_at"`
}

type UserMatchesResponse struct {
	Matches []UserJobMatchResponse `json:"matches"`
}
