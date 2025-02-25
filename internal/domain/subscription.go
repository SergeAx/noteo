package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID        uuid.UUID
	UserID    TelegramUserID
	ProjectID uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SubscriptionRepository interface {
	Create(subscription *Subscription) error
	Delete(userID TelegramUserID, projectID uuid.UUID) error
	GetByProject(projectID uuid.UUID) ([]*Subscription, error)
	GetByUser(userID TelegramUserID) ([]*Subscription, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Subscribe(userID TelegramUserID, projectID uuid.UUID) error {
	subscription := &Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(subscription); err != nil {
		return fmt.Errorf("creating subscription: %w", err)
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(userID TelegramUserID, projectID uuid.UUID) error {
	if err := s.repo.Delete(userID, projectID); err != nil {
		return fmt.Errorf("deleting subscription: %w", err)
	}
	return nil
}

func (s *SubscriptionService) GetProjectSubscriptions(projectID uuid.UUID) ([]*Subscription, error) {
	subscriptions, err := s.repo.GetByProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("getting project subscriptions: %w", err)
	}
	return subscriptions, nil
}

func (s *SubscriptionService) GetUserSubscriptions(userID TelegramUserID) ([]*Subscription, error) {
	subscriptions, err := s.repo.GetByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("getting user subscriptions: %w", err)
	}
	return subscriptions, nil
}
