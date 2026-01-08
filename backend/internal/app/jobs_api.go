package app

import (
	"net/http"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	jobhandler "github.com/jobping/backend/internal/features/job/handler"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
	jobsvc "github.com/jobping/backend/internal/features/job/service"
	notificationhandler "github.com/jobping/backend/internal/features/notification/handler"
	notificationrepo "github.com/jobping/backend/internal/features/notification/repository"
	notificationsvc "github.com/jobping/backend/internal/features/notification/service"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
	"github.com/jobping/backend/internal/server"
)

type JobsAPIApp struct {
	Router http.Handler
}

func BuildJobsAPI() (*JobsAPIApp, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build job feature dependencies
	jobRepo := jobrepo.NewJobRepository(db)
	// Create minimal service for GetJobs (only needs repo)
	jobService := jobsvc.NewJobService(jobRepo, nil, nil, nil)
	jobHandler := jobhandler.NewJobHandler(jobService)

	// 4. Build notification feature dependencies
	notifRepo := notificationrepo.NewNotificationRepository(db)
	userRepo := userrepo.NewUserRepository(db)
	matchRepo := userrepo.NewUserJobMatchRepository(db)
	notificationService := notificationsvc.NewNotificationService(jobRepo, userRepo, matchRepo, notifRepo)
	notificationHandler := notificationhandler.NewHTTPHandler(notificationService)

	// 5. Build router (job + notification routes)
	router := server.NewJobsRouter(jobHandler, notificationHandler)

	return &JobsAPIApp{
		Router: router,
	}, nil
}

