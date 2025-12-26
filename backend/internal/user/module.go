package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobping/backend/internal/user/handler"
	"github.com/jobping/backend/internal/user/repository"
	"github.com/jobping/backend/internal/user/service"
)

func RegisterRoutes(r chi.Router, db *pgxpool.Pool, jwtSecret string, jwtExpiryHours int) {
	userRepo := repository.NewUserRepository(db)
	prefRepo := repository.NewPreferenceRepository(db)
	svc := service.NewUserService(userRepo, prefRepo)
	auth := handler.NewAuthMiddleware(jwtSecret, jwtExpiryHours)
	h := handler.NewUserHandler(svc, auth)

	// Public routes
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Get("/preferences", h.GetPreferences)
		r.Post("/preferences", h.CreatePreference)
		r.Put("/preferences/{id}", h.UpdatePreference)
		r.Delete("/preferences/{id}", h.DeletePreference)
	})
}
