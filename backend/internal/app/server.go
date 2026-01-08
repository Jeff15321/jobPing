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
	userhandler "github.com/jobping/backend/internal/features/user/handler"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
	usersvc "github.com/jobping/backend/internal/features/user/service"
	"github.com/jobping/backend/internal/server"
)

type ServerApp struct {
	Router http.Handler
}

// BuildServer builds the combined app for local server development
func BuildServer() (*ServerApp, error) {
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
	jobService := jobsvc.NewJobService(jobRepo, nil, nil, nil)
	jobHandler := jobhandler.NewJobHandler(jobService)

	// 5. Build notification feature dependencies
	notifRepo := notificationrepo.NewNotificationRepository(db)
	notificationService := notificationsvc.NewNotificationService(jobRepo, userRepo, matchRepo, notifRepo)
	notificationHandler := notificationhandler.NewHTTPHandler(notificationService)

	// 6. Build router (combined for local dev)
	router := server.NewRouterWithNotification(userHandler, auth, jobHandler, notificationHandler)

	return &ServerApp{
		Router: router,
	}, nil
}

