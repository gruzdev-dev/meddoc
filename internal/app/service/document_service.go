package service

import (
	"context"

	"github.com/gruzdev-dev/meddoc/internal/domain"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentService struct {
	repo domain.DocumentRepository
}

func NewDocumentService(repo domain.DocumentRepository) *DocumentService {
	return &DocumentService{
		repo: repo,
	}
}

func (s *DocumentService) CreateDocument(ctx context.Context, doc *domain.Document) error {
	if doc.Title == "" {
		return domain.ErrEmptyTitle
	}
	if doc.Content == "" {
		return domain.ErrEmptyContent
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		logger.Error("failed to create document in repository", err)
		return domain.ErrDatabaseError
	}
	return nil
}

func (s *DocumentService) GetDocument(ctx context.Context, id primitive.ObjectID) (*domain.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get document from repository", err, "id", id)
		return nil, domain.ErrDatabaseError
	}
	if doc == nil {
		logger.Warn("document not found in repository", "id", id)
		return nil, domain.ErrDocumentNotFound
	}
	return doc, nil
}

func (s *DocumentService) GetAllDocuments(ctx context.Context) ([]*domain.Document, error) {
	docs, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("failed to get documents from repository", err)
		return nil, domain.ErrDatabaseError
	}
	return docs, nil
}

func (s *DocumentService) UpdateDocument(ctx context.Context, doc *domain.Document) error {
	if doc.Title == "" {
		return domain.ErrEmptyTitle
	}
	if doc.Content == "" {
		return domain.ErrEmptyContent
	}

	existing, err := s.repo.GetByID(ctx, doc.ID)
	if err != nil {
		logger.Error("failed to check document existence", err, "id", doc.ID)
		return domain.ErrDatabaseError
	}
	if existing == nil {
		logger.Warn("document not found for update", "id", doc.ID)
		return domain.ErrDocumentNotFound
	}

	if err := s.repo.Update(ctx, doc); err != nil {
		logger.Error("failed to update document in repository", err, "id", doc.ID)
		return domain.ErrDatabaseError
	}
	return nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id primitive.ObjectID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to check document existence", err, "id", id)
		return domain.ErrDatabaseError
	}
	if existing == nil {
		logger.Warn("document not found for deletion", "id", id)
		return domain.ErrDocumentNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("failed to delete document from repository", err, "id", id)
		return domain.ErrDatabaseError
	}
	return nil
}
