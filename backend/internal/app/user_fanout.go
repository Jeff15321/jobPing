package app

import (
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/database"
	userfanouthandler "github.com/jobping/backend/internal/features/user_fanout/handler"
	userfanoutsvc "github.com/jobping/backend/internal/features/user_fanout/service"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type UserFanoutApp struct {
	SQSHandler *userfanouthandler.SQSHandler
}

func BuildUserFanout() (*UserFanoutApp, error) {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 3. Build user fanout feature dependencies
	userRepo := userrepo.NewUserRepository(db)
	fanoutService := userfanoutsvc.NewFanoutService(userRepo)
	sqsHandler := userfanouthandler.NewSQSHandler(fanoutService)

	return &UserFanoutApp{
		SQSHandler: sqsHandler,
	}, nil
}


