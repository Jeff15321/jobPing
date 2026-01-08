package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobping/backend/internal/features/job/model"
)

type JobRepository interface {
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error)
	GetByURL(ctx context.Context, url string) (*model.Job, error)
	GetAll(ctx context.Context, limit int) ([]model.Job, error)
	GetProcessed(ctx context.Context, limit int) ([]model.Job, error)
	Update(ctx context.Context, job *model.Job) error
	UpdateCompanyInfo(ctx context.Context, id uuid.UUID, companyInfo map[string]interface{}) error
	ExistsByURL(ctx context.Context, url string) (bool, error)
	IsCompanyInfoFresh(ctx context.Context, id uuid.UUID) (bool, error)
}

type postgresJobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) JobRepository {
	return &postgresJobRepository{db: db}
}

func (r *postgresJobRepository) Create(ctx context.Context, job *model.Job) error {
	query := `
		INSERT INTO jobs (id, title, company, location, job_url, description, job_type, is_remote, min_salary, max_salary, date_posted, ai_score, ai_analysis, company_info, company_info_updated_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := r.db.Exec(ctx, query,
		job.ID, job.Title, job.Company, job.Location, job.JobURL, job.Description,
		job.JobType, job.IsRemote, job.MinSalary, job.MaxSalary, job.DatePosted,
		job.AIScore, job.AIAnalysis, job.CompanyInfo, job.CompanyInfoUpdatedAt, job.Status, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *postgresJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	query := `
		SELECT id, title, company, location, job_url, description, job_type, is_remote, min_salary, max_salary, date_posted, ai_score, ai_analysis, company_info, company_info_updated_at, status, created_at, updated_at
		FROM jobs WHERE id = $1
	`
	var job model.Job
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.Title, &job.Company, &job.Location, &job.JobURL, &job.Description,
		&job.JobType, &job.IsRemote, &job.MinSalary, &job.MaxSalary, &job.DatePosted,
		&job.AIScore, &job.AIAnalysis, &job.CompanyInfo, &job.CompanyInfoUpdatedAt, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *postgresJobRepository) GetByURL(ctx context.Context, url string) (*model.Job, error) {
	query := `
		SELECT id, title, company, location, job_url, description, job_type, is_remote, min_salary, max_salary, date_posted, ai_score, ai_analysis, company_info, company_info_updated_at, status, created_at, updated_at
		FROM jobs WHERE job_url = $1
	`
	var job model.Job
	err := r.db.QueryRow(ctx, query, url).Scan(
		&job.ID, &job.Title, &job.Company, &job.Location, &job.JobURL, &job.Description,
		&job.JobType, &job.IsRemote, &job.MinSalary, &job.MaxSalary, &job.DatePosted,
		&job.AIScore, &job.AIAnalysis, &job.CompanyInfo, &job.CompanyInfoUpdatedAt, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *postgresJobRepository) GetAll(ctx context.Context, limit int) ([]model.Job, error) {
	query := `
		SELECT id, title, company, location, job_url, description, job_type, is_remote, min_salary, max_salary, date_posted, ai_score, ai_analysis, company_info, company_info_updated_at, status, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
		LIMIT $1
	`
	return r.queryJobs(ctx, query, limit)
}

func (r *postgresJobRepository) GetProcessed(ctx context.Context, limit int) ([]model.Job, error) {
	query := `
		SELECT id, title, company, location, job_url, description, job_type, is_remote, min_salary, max_salary, date_posted, ai_score, ai_analysis, company_info, company_info_updated_at, status, created_at, updated_at
		FROM jobs
		WHERE status = 'processed'
		ORDER BY ai_score DESC NULLS LAST, created_at DESC
		LIMIT $1
	`
	return r.queryJobs(ctx, query, limit)
}

func (r *postgresJobRepository) queryJobs(ctx context.Context, query string, limit int) ([]model.Job, error) {
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []model.Job
	for rows.Next() {
		var job model.Job
		if err := rows.Scan(
			&job.ID, &job.Title, &job.Company, &job.Location, &job.JobURL, &job.Description,
			&job.JobType, &job.IsRemote, &job.MinSalary, &job.MaxSalary, &job.DatePosted,
			&job.AIScore, &job.AIAnalysis, &job.CompanyInfo, &job.CompanyInfoUpdatedAt, &job.Status, &job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *postgresJobRepository) UpdateCompanyInfo(ctx context.Context, id uuid.UUID, companyInfo map[string]interface{}) error {
	query := `UPDATE jobs SET company_info = $1, company_info_updated_at = NOW(), updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, companyInfo, id)
	return err
}

func (r *postgresJobRepository) Update(ctx context.Context, job *model.Job) error {
	query := `
		UPDATE jobs 
		SET ai_score = $1, ai_analysis = $2, status = $3, updated_at = $4 
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, job.AIScore, job.AIAnalysis, job.Status, job.UpdatedAt, job.ID)
	return err
}

func (r *postgresJobRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM jobs WHERE job_url = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, url).Scan(&exists)
	return exists, err
}

func (r *postgresJobRepository) IsCompanyInfoFresh(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `
		SELECT company_info_updated_at IS NOT NULL 
		AND company_info_updated_at > NOW() - INTERVAL '6 months'
		FROM jobs WHERE id = $1
	`
	var isFresh bool
	err := r.db.QueryRow(ctx, query, id).Scan(&isFresh)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isFresh, err
}


