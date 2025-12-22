package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/ai-job-scanner/internal/domain/job"
)

func UpsertJob(db *DB, j *job.Job) error {
	query := `
		INSERT INTO jobs (id, title, company, location, description, url, posted_at, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			company = EXCLUDED.company,
			location = EXCLUDED.location,
			description = EXCLUDED.description,
			url = EXCLUDED.url,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query,
		j.ID,
		j.Title,
		j.Company,
		j.Location,
		j.Description,
		j.URL,
		j.PostedAt,
		time.Now(),
	)

	return err
}

func GetJobs(db *DB, limit int) ([]*job.Job, error) {
	query := `
		SELECT id, title, company, location, description, url, posted_at, fetched_at, ai_analysis, created_at
		FROM jobs
		ORDER BY posted_at DESC
		LIMIT $1
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*job.Job
	for rows.Next() {
		j := &job.Job{}
		var aiAnalysisJSON sql.NullString

		err := rows.Scan(
			&j.ID,
			&j.Title,
			&j.Company,
			&j.Location,
			&j.Description,
			&j.URL,
			&j.PostedAt,
			&j.FetchedAt,
			&aiAnalysisJSON,
			&j.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if aiAnalysisJSON.Valid {
			json.Unmarshal([]byte(aiAnalysisJSON.String), &j.AIAnalysis)
		}

		jobs = append(jobs, j)
	}

	return jobs, nil
}

func GetJobByID(db *DB, id string) (*job.Job, error) {
	query := `
		SELECT id, title, company, location, description, url, posted_at, fetched_at, ai_analysis, created_at
		FROM jobs
		WHERE id = $1
	`

	j := &job.Job{}
	var aiAnalysisJSON sql.NullString

	err := db.QueryRow(query, id).Scan(
		&j.ID,
		&j.Title,
		&j.Company,
		&j.Location,
		&j.Description,
		&j.URL,
		&j.PostedAt,
		&j.FetchedAt,
		&aiAnalysisJSON,
		&j.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	if aiAnalysisJSON.Valid {
		json.Unmarshal([]byte(aiAnalysisJSON.String), &j.AIAnalysis)
	}

	return j, nil
}
