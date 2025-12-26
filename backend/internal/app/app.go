package app

import (
	"net/http"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	"github.com/jobping/backend/internal/features/user/handler"
	"github.com/jobping/backend/internal/features/user/repository"
	"github.com/jobping/backend/internal/features/user/service"
	"github.com/jobping/backend/internal/server"
)

type App struct {
	Router http.Handler
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
	userRepo := repository.NewUserRepository(db)
	prefRepo := repository.NewPreferenceRepository(db)
	userService := service.NewUserService(userRepo, prefRepo)
	auth := handler.NewAuthMiddleware(cfg.JWTSecret, cfg.JWTExpiry)
	userHandler := handler.NewUserHandler(userService, auth)

	// 4. Build router
	router := server.NewRouter(userHandler, auth)

	return &App{
		Router: router,
	}, nil
}
