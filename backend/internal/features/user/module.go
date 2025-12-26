package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/jobping/backend/internal/features/user/handler"
)

func RegisterRoutes(r chi.Router, userHandler *handler.UserHandler, auth *handler.AuthMiddleware) {
	// Public routes
	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Get("/preferences", userHandler.GetPreferences)
		r.Post("/preferences", userHandler.CreatePreference)
		r.Put("/preferences/{id}", userHandler.UpdatePreference)
		r.Delete("/preferences/{id}", userHandler.DeletePreference)
	})
}
