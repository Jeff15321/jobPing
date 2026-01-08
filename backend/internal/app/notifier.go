package app

import (
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
	notificationhandler "github.com/jobping/backend/internal/features/notification/handler"
	notificationrepo "github.com/jobping/backend/internal/features/notification/repository"
	notificationsvc "github.com/jobping/backend/internal/features/notification/service"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type NotifierApp struct {
	SQSHandler *notificationhandler.SQSHandler
}

func BuildNotifier() (*NotifierApp, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build notification feature dependencies
	jobRepo := jobrepo.NewJobRepository(db)
	userRepo := userrepo.NewUserRepository(db)
	matchRepo := userrepo.NewUserJobMatchRepository(db)
	notifRepo := notificationrepo.NewNotificationRepository(db)
	notificationService := notificationsvc.NewNotificationService(jobRepo, userRepo, matchRepo, notifRepo)
	sqsHandler := notificationhandler.NewSQSHandler(notificationService)

	return &NotifierApp{
		SQSHandler: sqsHandler,
	}, nil
}


