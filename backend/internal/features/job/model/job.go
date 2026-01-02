package model

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID
	Title       string
	Company     string
	Location    string
	JobURL      string
	Description string
	JobType     string
	IsRemote    bool
	MinSalary   *float64
	MaxSalary   *float64
	DatePosted  string
	AIScore     *int
	AIAnalysis  *string
	Status      JobStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusProcessed JobStatus = "processed"
	JobStatusFailed    JobStatus = "failed"
)

