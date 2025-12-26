package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobping/backend/internal/config"
	"github.com/jobping/backend/internal/features/users"
	"github.com/jobping/backend/internal/middleware"
)

func NewRouter(cfg *config.Config, db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize layers
	usersRepo := users.NewRepository(db)
	usersService := users.NewService(usersRepo, authMiddleware)
	usersHandler := users.NewHandler(usersService)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/register", usersHandler.Register)
		r.Post("/login", usersHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// Preferences CRUD
			r.Get("/preferences", usersHandler.GetPreferences)
			r.Post("/preferences", usersHandler.CreatePreference)
			r.Put("/preferences/{id}", usersHandler.UpdatePreference)
			r.Delete("/preferences/{id}", usersHandler.DeletePreference)
		})
	})

	return r
}

