package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	jobrepo "github.com/jobping/backend/internal/features/job/repository"
	notificationrepo "github.com/jobping/backend/internal/features/notification/repository"
	userrepo "github.com/jobping/backend/internal/features/user/repository"
)

type NotificationService struct {
	jobRepo    jobrepo.JobRepository
	userRepo   userrepo.UserRepository
	matchRepo  userrepo.UserJobMatchRepository
	notifRepo  notificationrepo.NotificationRepository
}

func NewNotificationService(
	jobRepo jobrepo.JobRepository,
	userRepo userrepo.UserRepository,
	matchRepo userrepo.UserJobMatchRepository,
	notifRepo notificationrepo.NotificationRepository,
) *NotificationService {
	return &NotificationService{
		jobRepo:   jobRepo,
		userRepo:  userRepo,
		matchRepo: matchRepo,
		notifRepo: notifRepo,
	}
}

// SendNotification creates a notification event for testing (stores in notifications table)
func (s *NotificationService) SendNotification(ctx context.Context, jobID, userID uuid.UUID) error {
	// Fetch user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		log.Printf("User not found: %s", userID)
		return nil
	}

	// Fetch job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Job not found: %s", jobID)
		return nil
	}

	// Fetch match
	match, err := s.matchRepo.GetByUserAndJob(ctx, userID, jobID)
	if err != nil {
		return err
	}
	if match == nil {
		log.Printf("Match not found for user %s and job %s", userID, jobID)
		return nil
	}

	// Create notification event
	notification := &notificationrepo.Notification{
		ID:            uuid.New(),
		UserID:        userID,
		JobID:         jobID,
		MatchID:       match.ID,
		JobTitle:      job.Title,
		Company:       job.Company,
		JobURL:        job.JobURL,
		MatchingScore: match.Score,
		AIAnalysis:    match.Analysis,
		CreatedAt:     time.Now(),
	}

	if err := s.notifRepo.Create(ctx, notification); err != nil {
		log.Printf("Failed to create notification: %v", err)
		return err
	}

	// Mark match as notified
	if err := s.matchRepo.MarkNotified(ctx, match.ID); err != nil {
		log.Printf("Failed to mark match as notified: %v", err)
		// Don't fail the whole operation
	}

	log.Printf("Created notification for user %s about job %s (score: %d)", user.Username, job.Title, match.Score)
	return nil
}

// GetNotifications returns notifications for a user (or all if userID is nil)
func (s *NotificationService) GetNotifications(ctx context.Context, userID *uuid.UUID, limit int) ([]notificationrepo.Notification, error) {
	if userID != nil {
		return s.notifRepo.GetByUserID(ctx, *userID, limit)
	}
	return s.notifRepo.GetAll(ctx, limit)
}

