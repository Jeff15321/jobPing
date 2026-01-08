package app

import (
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	jobanalysishandler "github.com/jobping/backend/internal/features/job_analysis/handler"
	jobanalysissvc "github.com/jobping/backend/internal/features/job_analysis/service"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
)

type JobAnalysisApp struct {
	SQSHandler *jobanalysishandler.SQSHandler
}

func BuildJobAnalysis() (*JobAnalysisApp, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build job analysis feature dependencies
	jobRepo := jobrepo.NewJobRepository(db)
	aiClient := jobanalysissvc.NewAIClient()
	jobAnalysisService := jobanalysissvc.NewJobAnalysisService(jobRepo, aiClient)
	sqsHandler := jobanalysishandler.NewSQSHandler(jobAnalysisService)

	return &JobAnalysisApp{
		SQSHandler: sqsHandler,
	}, nil
}


