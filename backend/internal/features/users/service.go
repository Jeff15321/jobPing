package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jobping/backend/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
	auth *middleware.AuthMiddleware
}

func NewService(repo *Repository, auth *middleware.AuthMiddleware) *Service {
	return &Service{repo: repo, auth: auth}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existing, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (*PreferencesResponse, error) {
	prefs, err := s.repo.GetPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := &PreferencesResponse{
		Preferences: make([]PreferenceResponse, len(prefs)),
	}
	for i, p := range prefs {
		response.Preferences[i] = PreferenceResponse{
			ID:    p.ID,
			Key:   p.Key,
			Value: p.Value,
		}
	}
	return response, nil
}

func (s *Service) CreatePreference(ctx context.Context, userID uuid.UUID, req CreatePreferenceRequest) (*PreferenceResponse, error) {
	// Check if preference with this key exists
	existing, err := s.repo.GetPreferenceByUserIDAndKey(ctx, userID, req.Key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPreferenceExists
	}

	now := time.Now()
	pref := &Preference{
		ID:        uuid.New(),
		UserID:    userID,
		Key:       req.Key,
		Value:     req.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreatePreference(ctx, pref); err != nil {
		return nil, err
	}

	return &PreferenceResponse{
		ID:    pref.ID,
		Key:   pref.Key,
		Value: pref.Value,
	}, nil
}

func (s *Service) UpdatePreference(ctx context.Context, userID, prefID uuid.UUID, req UpdatePreferenceRequest) (*PreferenceResponse, error) {
	pref, err := s.repo.GetPreferenceByID(ctx, prefID)
	if err != nil {
		return nil, err
	}
	if pref == nil || pref.UserID != userID {
		return nil, ErrPreferenceNotFound
	}

	pref.Value = req.Value
	pref.UpdatedAt = time.Now()

	if err := s.repo.UpdatePreference(ctx, pref); err != nil {
		return nil, err
	}

	return &PreferenceResponse{
		ID:    pref.ID,
		Key:   pref.Key,
		Value: pref.Value,
	}, nil
}

func (s *Service) DeletePreference(ctx context.Context, userID, prefID uuid.UUID) error {
	pref, err := s.repo.GetPreferenceByID(ctx, prefID)
	if err != nil {
		return err
	}
	if pref == nil || pref.UserID != userID {
		return ErrPreferenceNotFound
	}

	return s.repo.DeletePreference(ctx, prefID)
}

