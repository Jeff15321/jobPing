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
	UpdateAIPrompt(ctx context.Context, userID uuid.UUID, prompt string) error
	UpdateDiscordWebhook(ctx context.Context, userID uuid.UUID, webhook string) error
	UpdateNotifyThreshold(ctx context.Context, userID uuid.UUID, threshold int) error
	GetUsersWithPrompts(ctx context.Context) ([]model.User, error)
}

type UserJobMatchRepository interface {
	Create(ctx context.Context, match *model.UserJobMatch) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserJobMatch, error)
	GetByUserAndJob(ctx context.Context, userID, jobID uuid.UUID) (*model.UserJobMatch, error)
	MarkNotified(ctx context.Context, id uuid.UUID) error
	GetUnnotifiedAboveThreshold(ctx context.Context) ([]model.UserJobMatch, error)
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
		INSERT INTO users (id, username, password_hash, notify_threshold, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	threshold := user.NotifyThreshold
	if threshold == 0 {
		threshold = 70
	}
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, threshold, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, ai_prompt, discord_webhook, COALESCE(notify_threshold, 70), created_at, updated_at
		FROM users WHERE username = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.AIPrompt, &user.DiscordWebhook, &user.NotifyThreshold, &user.CreatedAt, &user.UpdatedAt,
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
		SELECT id, username, password_hash, ai_prompt, discord_webhook, COALESCE(notify_threshold, 70), created_at, updated_at
		FROM users WHERE id = $1
	`
	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.AIPrompt, &user.DiscordWebhook, &user.NotifyThreshold, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresUserRepository) UpdateAIPrompt(ctx context.Context, userID uuid.UUID, prompt string) error {
	query := `UPDATE users SET ai_prompt = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, prompt, userID)
	return err
}

func (r *postgresUserRepository) UpdateDiscordWebhook(ctx context.Context, userID uuid.UUID, webhook string) error {
	query := `UPDATE users SET discord_webhook = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, webhook, userID)
	return err
}

func (r *postgresUserRepository) UpdateNotifyThreshold(ctx context.Context, userID uuid.UUID, threshold int) error {
	query := `UPDATE users SET notify_threshold = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, threshold, userID)
	return err
}

func (r *postgresUserRepository) GetUsersWithPrompts(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, username, password_hash, ai_prompt, discord_webhook, COALESCE(notify_threshold, 70), created_at, updated_at
		FROM users WHERE ai_prompt IS NOT NULL AND ai_prompt != ''
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.AIPrompt, &user.DiscordWebhook, &user.NotifyThreshold, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
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

// UserJobMatch Repository

type postgresUserJobMatchRepository struct {
	db *pgxpool.Pool
}

func NewUserJobMatchRepository(db *pgxpool.Pool) UserJobMatchRepository {
	return &postgresUserJobMatchRepository{db: db}
}

func (r *postgresUserJobMatchRepository) Create(ctx context.Context, match *model.UserJobMatch) error {
	query := `
		INSERT INTO user_job_matches (id, user_id, job_id, score, analysis, notified, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, job_id) DO UPDATE SET score = $4, analysis = $5
	`
	_, err := r.db.Exec(ctx, query,
		match.ID, match.UserID, match.JobID, match.Score, match.Analysis, match.Notified, match.CreatedAt,
	)
	return err
}

func (r *postgresUserJobMatchRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.UserJobMatch, error) {
	query := `
		SELECT id, user_id, job_id, score, analysis, notified, created_at
		FROM user_job_matches WHERE user_id = $1
		ORDER BY score DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []model.UserJobMatch
	for rows.Next() {
		var match model.UserJobMatch
		if err := rows.Scan(
			&match.ID, &match.UserID, &match.JobID, &match.Score, &match.Analysis, &match.Notified, &match.CreatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, rows.Err()
}

func (r *postgresUserJobMatchRepository) GetByUserAndJob(ctx context.Context, userID, jobID uuid.UUID) (*model.UserJobMatch, error) {
	query := `
		SELECT id, user_id, job_id, score, analysis, notified, created_at
		FROM user_job_matches WHERE user_id = $1 AND job_id = $2
	`
	var match model.UserJobMatch
	err := r.db.QueryRow(ctx, query, userID, jobID).Scan(
		&match.ID, &match.UserID, &match.JobID, &match.Score, &match.Analysis, &match.Notified, &match.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *postgresUserJobMatchRepository) MarkNotified(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE user_job_matches SET notified = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresUserJobMatchRepository) GetUnnotifiedAboveThreshold(ctx context.Context) ([]model.UserJobMatch, error) {
	query := `
		SELECT m.id, m.user_id, m.job_id, m.score, m.analysis, m.notified, m.created_at
		FROM user_job_matches m
		JOIN users u ON m.user_id = u.id
		WHERE m.notified = FALSE AND m.score >= u.notify_threshold
		ORDER BY m.created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []model.UserJobMatch
	for rows.Next() {
		var match model.UserJobMatch
		if err := rows.Scan(
			&match.ID, &match.UserID, &match.JobID, &match.Score, &match.Analysis, &match.Notified, &match.CreatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, rows.Err()
}
