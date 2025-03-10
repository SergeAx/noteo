package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gitlab.com/trum/noteo/internal/domain"
)

type subscription struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID      domain.TelegramUserID
	ProjectID   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Muted       bool
	PausedUntil *time.Time
}

func (s *subscription) toDomain() *domain.Subscription {
	return &domain.Subscription{
		ID:          s.ID,
		UserID:      s.UserID,
		ProjectID:   s.ProjectID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Muted:       s.Muted,
		PausedUntil: s.PausedUntil,
	}
}

func subscriptionFromDomain(s *domain.Subscription) *subscription {
	return &subscription{
		ID:          s.ID,
		UserID:      s.UserID,
		ProjectID:   s.ProjectID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Muted:       s.Muted,
		PausedUntil: s.PausedUntil,
	}
}

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(subscription *domain.Subscription) error {
	dbSubscription := subscriptionFromDomain(subscription)
	if err := r.db.Create(dbSubscription).Error; err != nil {
		return fmt.Errorf("creating subscription in db: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) Delete(userID domain.TelegramUserID, projectID uuid.UUID) error {
	if err := r.db.Where("user_id = ? AND project_id = ?", userID, projectID).Delete(&subscription{}).Error; err != nil {
		return fmt.Errorf("deleting subscription from db: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByProject(projectID uuid.UUID) ([]*domain.Subscription, error) {
	var subscriptions []subscription
	if err := r.db.Where("project_id = ?", projectID).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("getting project subscriptions from db: %w", err)
	}

	result := make([]*domain.Subscription, len(subscriptions))
	for i := range subscriptions {
		result[i] = subscriptions[i].toDomain()
	}
	return result, nil
}

func (r *SubscriptionRepository) GetByUser(userID domain.TelegramUserID) ([]*domain.Subscription, error) {
	var subscriptions []subscription
	if err := r.db.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("getting user subscriptions from db: %w", err)
	}

	result := make([]*domain.Subscription, len(subscriptions))
	for i := range subscriptions {
		result[i] = subscriptions[i].toDomain()
	}
	return result, nil
}

func (r *SubscriptionRepository) Update(subscription *domain.Subscription) error {
	dbSubscription := subscriptionFromDomain(subscription)
	if err := r.db.Save(dbSubscription).Error; err != nil {
		return fmt.Errorf("updating subscription in db: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByUserAndProject(userID domain.TelegramUserID, projectID uuid.UUID) (*domain.Subscription, error) {
	var sub subscription
	if err := r.db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&sub).Error; err != nil {
		return nil, fmt.Errorf("getting subscription from db: %w", err)
	}
	return sub.toDomain(), nil
}
