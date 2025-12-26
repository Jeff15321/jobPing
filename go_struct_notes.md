project/
├── main.go
├── entity.go
├── handler.go
├── service.go
├── repository.go
├── config.go
├── go.mod

// go.mod
module example.com/project

go 1.22

// entity.go
package main

type User struct {
	ID   int
	Name string
}

// repository.go
package main

type UserRepository interface {
	GetAll() []User
}

type InMemoryUserRepository struct{}

func NewUserRepository() UserRepository {
	return &InMemoryUserRepository{}
}

func (r *InMemoryUserRepository) GetAll() []User {
	return []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
}

// service.go
package main

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUsers() []User {
	return s.repo.GetAll()
}

// handler.go
package main

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := h.service.GetUsers()
	json.NewEncoder(w).Encode(users)
}

// config.go
package main

import "os"

type Config struct {
	Port string
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port}
}

// main.go
package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := LoadConfig()

	repo := NewUserRepository()
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	http.HandleFunc("/users", handler.GetUsers)

	log.Println("Listening on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}





====================================================================================

# Feature driven architecture:

internal/
├── user/
│   ├── handler.go        # HTTP / transport layer
│   ├── service.go        # Use cases / business logic
│   ├── repository.go    # Persistence (interfaces + impl)
│   ├── model.go         # Domain entities
│   └── errors.go        # Domain-specific errors
│
├── product/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── model.go
│
├── shared/
│   ├── database.go      # DB connection
│   ├── logger.go
│   └── pagination.go
│
└── config/
    └── config.go

# when one feature is very complex:
internal/
├── user/
│   ├── handler/
│   │   ├── http.go
│   │   ├── middleware.go
│   │   └── dto.go
│   │
│   ├── service/
│   │   ├── user_service.go
│   │   ├── auth_service.go
│   │   └── permission_service.go
│   │
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── session_repository.go
│   │   └── cache_repository.go
│   │
│   ├── model/
│   │   ├── user.go
│   │   ├── role.go
│   │   └── permission.go
│   │
│   ├── errors.go
│   └── module.go

starts at main.go, creates mux object for handling incoming http request
calls registerRoutes to set up  repo, svc, and handler, define protected as a http.HandlerFunc that is just the middleware function that calls the handler function 
the middleware function exists in user/handler/middleware.go, it checks the header and pass in any more information to the request body by using context
the handler function exists in user/handler/http.go, it then calls service and returns response as UserProfileResponse from dto.go (data transfer object)
the service function exists in user/service/user_service.go, it calls from the repo (remember to use interface instead of the concrete repo type) and applies business rules, return with error variable as none nil if there is an error

example:
project/
├── go.mod
├── main.go
└── internal/
    └── user/
        ├── handler/
        │   ├── http.go
        │   ├── middleware.go
        │   └── dto.go
        ├── service/
        │   └── user_service.go
        ├── repository/
        │   └── user_repository.go
        ├── model/
        │   └── user.go
        ├── errors.go
        └── module.go


// go.mod
module example.com/project

go 1.22


// main.go
package main

import (
	"log"
	"net/http"

	"example.com/project/internal/user"
)

func main() {
	mux := http.NewServeMux()
	user.RegisterRoutes(mux)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}


// internal/user/model/user.go
package model

type User struct {
	ID       string
	Name     string
	Email    string
	IsActive bool
}


// internal/user/errors.go
package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserInactive = errors.New("user is inactive")
)


// internal/user/repository/user_repository.go
package repository

import "example.com/project/internal/user/model"

type UserRepository interface {
	FindByID(id string) (*model.User, error)
}

type InMemoryUserRepository struct{}

func NewUserRepository() UserRepository {
	return &InMemoryUserRepository{}
}

func (r *InMemoryUserRepository) FindByID(id string) (*model.User, error) {
	if id != "123" {
		return nil, nil
	}

	return &model.User{
		ID:       "123",
		Name:     "Alice",
		Email:    "alice@example.com",
		IsActive: true,
	}, nil
}


// internal/user/service/user_service.go
package service

import (
	"example.com/project/internal/user"
	"example.com/project/internal/user/model"
	"example.com/project/internal/user/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(userID string) (*model.User, error) {
	userEntity, err := s.repo.FindByID(userID)
	if err != nil || userEntity == nil {
		return nil, user.ErrUserNotFound
	}

	if !userEntity.IsActive {
		return nil, user.ErrUserInactive
	}

	return userEntity, nil
}


// internal/user/handler/dto.go
package handler

type UserProfileResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}


// internal/user/handler/middleware.go
package handler

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) string {
	return ctx.Value(userIDKey).(string)
}


// internal/user/handler/http.go
package handler

import (
	"encoding/json"
	"net/http"

	"example.com/project/internal/user"
	"example.com/project/internal/user/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	userEntity, err := h.service.GetProfile(userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == user.ErrUserInactive {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := UserProfileResponse{
		ID:    userEntity.ID,
		Name:  userEntity.Name,
		Email: userEntity.Email,
	}

	json.NewEncoder(w).Encode(resp)
}


// internal/user/module.go
package user

import (
	"net/http"

	"example.com/project/internal/user/handler"
	"example.com/project/internal/user/repository"
	"example.com/project/internal/user/service"
)

func RegisterRoutes(mux *http.ServeMux) {
	repo := repository.NewUserRepository()
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	protected := handler.AuthMiddleware(http.HandlerFunc(h.GetProfile))
	mux.Handle("/users/profile", protected)
}
