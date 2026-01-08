package app

import (
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
	useranalysishandler "github.com/jobping/backend/internal/features/user_analysis/handler"
	useranalysissvc "github.com/jobping/backend/internal/features/user_analysis/service"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type UserAnalysisApp struct {
	SQSHandler *useranalysishandler.SQSHandler
}

func BuildUserAnalysis() (*UserAnalysisApp, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build user analysis feature dependencies
	jobRepo := jobrepo.NewJobRepository(db)
	userRepo := userrepo.NewUserRepository(db)
	matchRepo := userrepo.NewUserJobMatchRepository(db)
	aiClient := useranalysissvc.NewAIClient()
	userAnalysisService := useranalysissvc.NewUserAnalysisService(jobRepo, userRepo, matchRepo, aiClient)
	sqsHandler := useranalysishandler.NewSQSHandler(userAnalysisService)

	return &UserAnalysisApp{
		SQSHandler: sqsHandler,
	}, nil
}


