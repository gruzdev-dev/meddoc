package document

import (
	"context"
	"errors"
	"time"

	"github.com/gruzdev-dev/meddoc/app/models"
)

var (
	ErrAccessDenied = errors.New("access denied")
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id string) (*models.Document, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Document, error)
}

type Service struct {
	repo DocumentRepository
}

func NewService(repo DocumentRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateDocument(ctx context.Context, data models.DocumentCreation, userID string) error {
	doc := &models.Document{
		Title:       data.Title,
		Description: data.Description,
		Date:        data.Date,
		File:        data.File,
		Category:    data.Category,
		Priority:    data.Priority,
		Content:     data.Content,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.repo.Create(ctx, doc)
}

func (s *Service) GetDocument(ctx context.Context, id string, userID string) (*models.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if doc.UserID != userID {
		return nil, ErrAccessDenied
	}

	return doc, nil
}

func (s *Service) GetUserDocuments(ctx context.Context, userID string) ([]*models.Document, error) {
	return s.repo.GetByUserID(ctx, userID)
}
