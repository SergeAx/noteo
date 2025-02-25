package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	Name        string
	Token       string
	PublisherID TelegramUserID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectRepository interface {
	Create(project *Project) error
	GetByID(id uuid.UUID) (*Project, error)
	GetByToken(token string) (*Project, error)
	GetByPublisher(publisherID TelegramUserID) ([]*Project, error)
	UpdateName(id uuid.UUID, name string) error
	UpdateToken(id uuid.UUID, token string) error
}

type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (s *ProjectService) Create(publisherID TelegramUserID, name string) (*Project, error) {
	project := &Project{
		ID:          uuid.New(),
		Name:        name,
		Token:       uuid.New().String(),
		PublisherID: publisherID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(project); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	return project, nil
}

func (s *ProjectService) GetByID(id uuid.UUID) (*Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("getting project by id: %w", err)
	}
	return project, nil
}

func (s *ProjectService) GetByToken(token string) (*Project, error) {
	project, err := s.repo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("getting project by token: %w", err)
	}
	return project, nil
}

func (s *ProjectService) GetByPublisher(publisherID TelegramUserID) ([]*Project, error) {
	projects, err := s.repo.GetByPublisher(publisherID)
	if err != nil {
		return nil, fmt.Errorf("getting projects by publisher: %w", err)
	}
	return projects, nil
}

func (s *ProjectService) UpdateName(id uuid.UUID, name string) error {
	if err := s.repo.UpdateName(id, name); err != nil {
		return fmt.Errorf("updating project name: %w", err)
	}
	return nil
}

func (s *ProjectService) RegenerateToken(id uuid.UUID) (string, error) {
	token := uuid.New().String()
	if err := s.repo.UpdateToken(id, token); err != nil {
		return "", fmt.Errorf("regenerating project token: %w", err)
	}
	return token, nil
}
