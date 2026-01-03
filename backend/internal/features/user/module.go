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

		// Profile and settings
		r.Get("/me", userHandler.GetProfile)
		r.Put("/me/prompt", userHandler.UpdatePrompt)
		r.Put("/me/discord", userHandler.UpdateDiscord)
		r.Put("/me/threshold", userHandler.UpdateThreshold)
		r.Get("/me/matches", userHandler.GetMatches)

		// Preferences (legacy)
		r.Get("/preferences", userHandler.GetPreferences)
		r.Post("/preferences", userHandler.CreatePreference)
		r.Put("/preferences/{id}", userHandler.UpdatePreference)
		r.Delete("/preferences/{id}", userHandler.DeletePreference)
	})
}
