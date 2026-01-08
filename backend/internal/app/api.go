package app

import (
	"net/http"

	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	userhandler "github.com/jobping/backend/internal/features/user/handler"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
	usersvc "github.com/jobping/backend/internal/features/user/service"
	"github.com/jobping/backend/internal/server"
)

type APIApp struct {
	Router http.Handler
}

func BuildAPI() (*APIApp, error) {
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

	// 4. Build router (user routes only)
	router := server.NewUserRouter(userHandler, auth)

	return &APIApp{
		Router: router,
	}, nil
}


