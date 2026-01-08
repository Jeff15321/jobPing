package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	JobID         uuid.UUID
	MatchID       uuid.UUID
	JobTitle      string
	Company       string
	JobURL        string
	MatchingScore int
	AIAnalysis    map[string]interface{}
	CreatedAt     time.Time
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error)
	GetAll(ctx context.Context, limit int) ([]Notification, error)
}

type postgresNotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &postgresNotificationRepository{db: db}
}

func (r *postgresNotificationRepository) Create(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, job_id, match_id, job_title, company, job_url, matching_score, ai_analysis, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		notification.ID, notification.UserID, notification.JobID, notification.MatchID,
		notification.JobTitle, notification.Company, notification.JobURL,
		notification.MatchingScore, notification.AIAnalysis, notification.CreatedAt,
	)
	return err
}

func (r *postgresNotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Notification, error) {
	query := `
		SELECT id, user_id, job_id, match_id, job_title, company, job_url, matching_score, ai_analysis, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	return r.queryNotifications(ctx, query, userID, limit)
}

func (r *postgresNotificationRepository) GetAll(ctx context.Context, limit int) ([]Notification, error) {
	query := `
		SELECT id, user_id, job_id, match_id, job_title, company, job_url, matching_score, ai_analysis, created_at
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1
	`
	return r.queryNotifications(ctx, query, limit)
}

func (r *postgresNotificationRepository) queryNotifications(ctx context.Context, query string, args ...interface{}) ([]Notification, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.JobID, &n.MatchID,
			&n.JobTitle, &n.Company, &n.JobURL, &n.MatchingScore,
			&n.AIAnalysis, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

