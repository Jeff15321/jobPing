package job

import "time"

type Job struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Company     string                 `json:"company"`
	Location    string                 `json:"location"`
	Description string                 `json:"description"`
	URL         string                 `json:"url"`
	PostedAt    time.Time              `json:"posted_at"`
	FetchedAt   time.Time              `json:"fetched_at"`
	AIAnalysis  map[string]interface{} `json:"ai_analysis,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type AIAnalysis struct {
	CompanyReputation string   `json:"company_reputation"`
	Benefits          []string `json:"benefits"`
	WorkCulture       string   `json:"work_culture"`
	Perks             []string `json:"perks"`
	Summary           string   `json:"summary"`
}
