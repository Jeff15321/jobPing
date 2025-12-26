package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobping/backend/internal/features/user/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type PreferenceRepository interface {
	Create(ctx context.Context, pref *model.Preference) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Preference, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Preference, error)
	GetByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*model.Preference, error)
	Update(ctx context.Context, pref *model.Preference) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users WHERE username = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type postgresPreferenceRepository struct {
	db *pgxpool.Pool
}

func NewPreferenceRepository(db *pgxpool.Pool) PreferenceRepository {
	return &postgresPreferenceRepository{db: db}
}

func (r *postgresPreferenceRepository) Create(ctx context.Context, pref *model.Preference) error {
	query := `
		INSERT INTO preferences (id, user_id, key, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		pref.ID, pref.UserID, pref.Key, pref.Value, pref.CreatedAt, pref.UpdatedAt,
	)
	return err
}

func (r *postgresPreferenceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Preference, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at
		FROM preferences WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []model.Preference
	for rows.Next() {
		var pref model.Preference
		if err := rows.Scan(
			&pref.ID, &pref.UserID, &pref.Key, &pref.Value, &pref.CreatedAt, &pref.UpdatedAt,
		); err != nil {
			return nil, err
		}
		prefs = append(prefs, pref)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prefs, nil
}

func (r *postgresPreferenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Preference, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at
		FROM preferences WHERE id = $1
	`
	var pref model.Preference
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pref.ID, &pref.UserID, &pref.Key, &pref.Value, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *postgresPreferenceRepository) GetByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*model.Preference, error) {
	query := `
		SELECT id, user_id, key, value, created_at, updated_at
		FROM preferences WHERE user_id = $1 AND key = $2
	`
	var pref model.Preference
	err := r.db.QueryRow(ctx, query, userID, key).Scan(
		&pref.ID, &pref.UserID, &pref.Key, &pref.Value, &pref.CreatedAt, &pref.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *postgresPreferenceRepository) Update(ctx context.Context, pref *model.Preference) error {
	query := `UPDATE preferences SET value = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, pref.Value, pref.UpdatedAt, pref.ID)
	return err
}

func (r *postgresPreferenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM preferences WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
