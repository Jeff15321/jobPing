package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func RunMigrations(db *DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS jobs (
			id VARCHAR(255) PRIMARY KEY,
			title VARCHAR(500) NOT NULL,
			company VARCHAR(255) NOT NULL,
			location VARCHAR(255),
			description TEXT,
			url VARCHAR(1000),
			posted_at TIMESTAMP,
			fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ai_analysis JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			preferences TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_job_matches (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			job_id VARCHAR(255) REFERENCES jobs(id),
			match_score FLOAT,
			notified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_posted_at ON jobs(posted_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_fetched_at ON jobs(fetched_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_user_job_matches_user_id ON user_job_matches(user_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
