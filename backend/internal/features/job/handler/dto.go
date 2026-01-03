package handler

import (
	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/job/model"
)

type JobResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Company    string    `json:"company"`
	Location   string    `json:"location"`
	JobURL     string    `json:"job_url"`
	JobType    string    `json:"job_type"`
	IsRemote   bool      `json:"is_remote"`
	MinSalary  *float64  `json:"min_salary,omitempty"`
	MaxSalary  *float64  `json:"max_salary,omitempty"`
	DatePosted string    `json:"date_posted,omitempty"`
	AIScore    *int      `json:"ai_score,omitempty"`
	AIAnalysis *string   `json:"ai_analysis,omitempty"`
}

type JobsResponse struct {
	Jobs []JobResponse `json:"jobs"`
}

func ToJobResponse(job model.Job) JobResponse {
	return JobResponse{
		ID:         job.ID,
		Title:      job.Title,
		Company:    job.Company,
		Location:   job.Location,
		JobURL:     job.JobURL,
		JobType:    job.JobType,
		IsRemote:   job.IsRemote,
		MinSalary:  job.MinSalary,
		MaxSalary:  job.MaxSalary,
		DatePosted: job.DatePosted,
		AIScore:    job.AIScore,
		AIAnalysis: job.AIAnalysis,
	}
}

func ToJobsResponse(jobs []model.Job) JobsResponse {
	response := JobsResponse{Jobs: make([]JobResponse, len(jobs))}
	for i, job := range jobs {
		response.Jobs[i] = ToJobResponse(job)
	}
	return response
}


