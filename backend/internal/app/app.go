package app

import (
	"net/http"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	jobhandler "github.com/jobping/backend/internal/features/job/handler"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
	jobsvc "github.com/jobping/backend/internal/features/job/service"
	userhandler "github.com/jobping/backend/internal/features/user/handler"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
	usersvc "github.com/jobping/backend/internal/features/user/service"
	"github.com/jobping/backend/internal/server"
)

type App struct {
	Router     http.Handler
	SQSHandler *jobhandler.SQSHandler
}

func Build() (*App, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build user feature dependencies
	userRepo := userrepo.NewUserRepository(db)
	prefRepo := userrepo.NewPreferenceRepository(db)
	matchRepo := userrepo.NewUserJobMatchRepository(db)
	userService := usersvc.NewUserService(userRepo, prefRepo, matchRepo)
	auth := userhandler.NewAuthMiddleware(cfg.JWTSecret, cfg.JWTExpiry)
	userHandler := userhandler.NewUserHandler(userService, auth)

	// 4. Build job feature dependencies
	jobRepo := jobrepo.NewJobRepository(db)
	aiClient := jobsvc.NewOpenAIClient()
	jobService := jobsvc.NewJobService(jobRepo, aiClient, userRepo, matchRepo)
	jobHandler := jobhandler.NewJobHandler(jobService)
	sqsHandler := jobhandler.NewSQSHandler(jobService)

	// 5. Build router
	router := server.NewRouter(userHandler, auth, jobHandler)

	return &App{
		Router:     router,
		SQSHandler: sqsHandler,
	}, nil
}
