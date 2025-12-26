package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jobping/backend/internal/features/user/model"
	"github.com/jobping/backend/internal/features/user/repository"
	"github.com/jobping/backend/internal/features/user/usererr"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
	prefRepo repository.PreferenceRepository
}

func NewUserService(userRepo repository.UserRepository, prefRepo repository.PreferenceRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		prefRepo: prefRepo,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	existing, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, usererr.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newUser := &model.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	userEntity, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if userEntity == nil {
		return nil, usererr.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(password)); err != nil {
		return nil, usererr.ErrInvalidCredentials
	}

	return userEntity, nil
}

func (s *UserService) GetPreferences(ctx context.Context, userID uuid.UUID) ([]model.Preference, error) {
	return s.prefRepo.GetByUserID(ctx, userID)
}

func (s *UserService) CreatePreference(ctx context.Context, userID uuid.UUID, key, value string) (*model.Preference, error) {
	existing, err := s.prefRepo.GetByUserIDAndKey(ctx, userID, key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, usererr.ErrPreferenceExists
	}

	now := time.Now()
	pref := &model.Preference{
		ID:        uuid.New(),
		UserID:    userID,
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.prefRepo.Create(ctx, pref); err != nil {
		return nil, err
	}

	return pref, nil
}

func (s *UserService) UpdatePreference(ctx context.Context, userID, prefID uuid.UUID, value string) (*model.Preference, error) {
	pref, err := s.prefRepo.GetByID(ctx, prefID)
	if err != nil {
		return nil, err
	}
	if pref == nil || pref.UserID != userID {
		return nil, usererr.ErrPreferenceNotFound
	}

	pref.Value = value
	pref.UpdatedAt = time.Now()

	if err := s.prefRepo.Update(ctx, pref); err != nil {
		return nil, err
	}

	return pref, nil
}

func (s *UserService) DeletePreference(ctx context.Context, userID, prefID uuid.UUID) error {
	pref, err := s.prefRepo.GetByID(ctx, prefID)
	if err != nil {
		return err
	}
	if pref == nil || pref.UserID != userID {
		return usererr.ErrPreferenceNotFound
	}

	return s.prefRepo.Delete(ctx, prefID)
}
