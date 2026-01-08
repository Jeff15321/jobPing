package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jobping/backend/internal/features/job"
	jobhandler "github.com/jobping/backend/internal/features/job/handler"
	notificationhandler "github.com/jobping/backend/internal/features/notification/handler"
	"github.com/jobping/backend/internal/features/user"
	userhandler "github.com/jobping/backend/internal/features/user/handler"
)

func NewRouter(userHandler *userhandler.UserHandler, auth *userhandler.AuthMiddleware, jobHandler *jobhandler.JobHandler) *chi.Mux {
	return NewRouterWithNotification(userHandler, auth, jobHandler, nil)
}

func NewRouterWithNotification(userHandler *userhandler.UserHandler, auth *userhandler.AuthMiddleware, jobHandler *jobhandler.JobHandler, notificationHandler *notificationhandler.HTTPHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		user.RegisterRoutes(r, userHandler, auth)
		job.RegisterRoutes(r, jobHandler)
		if notificationHandler != nil {
			r.Get("/notifications", notificationHandler.GetNotifications)
		}
	})

	return r
}

func NewUserRouter(userHandler *userhandler.UserHandler, auth *userhandler.AuthMiddleware) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		user.RegisterRoutes(r, userHandler, auth)
	})

	return r
}

func NewJobsRouter(jobHandler *jobhandler.JobHandler, notificationHandler *notificationhandler.HTTPHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/jobs", jobHandler.GetJobs)
		r.Get("/notifications", notificationHandler.GetNotifications)
	})

	return r
}
