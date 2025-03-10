package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sergeax/noteo/internal/domain"
)

type project struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name        string    `gorm:"uniqueIndex:idx_publisher_project_name"`
	Token       string    `gorm:"unique"`
	PublisherID domain.TelegramUserID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *project) toDomain() *domain.Project {
	return &domain.Project{
		ID:          p.ID,
		Name:        p.Name,
		Token:       p.Token,
		PublisherID: p.PublisherID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func projectFromDomain(p *domain.Project) *project {
	return &project{
		ID:          p.ID,
		Name:        p.Name,
		Token:       p.Token,
		PublisherID: p.PublisherID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) Create(project *domain.Project) error {
	dbProject := projectFromDomain(project)
	if err := r.db.Create(dbProject).Error; err != nil {
		return fmt.Errorf("creating project in db: %w", err)
	}
	return nil
}

func (r *ProjectRepository) GetByID(id uuid.UUID) (*domain.Project, error) {
	var project project
	if err := r.db.First(&project, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("getting project by id from db: %w", err)
	}
	return project.toDomain(), nil
}

func (r *ProjectRepository) GetByToken(token string) (*domain.Project, error) {
	var project project
	if err := r.db.First(&project, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("getting project by token from db: %w", err)
	}
	return project.toDomain(), nil
}

func (r *ProjectRepository) GetByPublisher(publisherID domain.TelegramUserID) ([]*domain.Project, error) {
	var projects []project
	if err := r.db.Where("publisher_id = ?", publisherID).Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("getting projects from db: %w", err)
	}

	result := make([]*domain.Project, len(projects))
	for i := range projects {
		result[i] = projects[i].toDomain()
	}
	return result, nil
}

func (r *ProjectRepository) UpdateName(id uuid.UUID, name string) error {
	if err := r.db.Model(&project{}).Where("id = ?", id).Update("name", name).Error; err != nil {
		return fmt.Errorf("updating project name in db: %w", err)
	}
	return nil
}

func (r *ProjectRepository) UpdateToken(id uuid.UUID, token string) error {
	if err := r.db.Model(&project{}).Where("id = ?", id).Update("token", token).Error; err != nil {
		return fmt.Errorf("updating project token in db: %w", err)
	}
	return nil
}
