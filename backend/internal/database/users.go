package database

import (
	"fmt"

	"github.com/yourusername/ai-job-scanner/internal/domain/user"
)

func CreateUser(db *DB, email, preferences string) (*user.User, error) {
	query := `
		INSERT INTO users (email, preferences)
		VALUES ($1, $2)
		RETURNING id, email, preferences, created_at, updated_at
	`

	u := &user.User{}
	err := db.QueryRow(query, email, preferences).Scan(
		&u.ID,
		&u.Email,
		&u.Preferences,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

func GetUserByEmail(db *DB, email string) (*user.User, error) {
	query := `
		SELECT id, email, preferences, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u := &user.User{}
	err := db.QueryRow(query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Preferences,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return u, nil
}

func UpdateUserPreferences(db *DB, userID int, preferences string) error {
	query := `
		UPDATE users
		SET preferences = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	_, err := db.Exec(query, preferences, userID)
	return err
}
