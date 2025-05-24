package document

import (
	"context"

	"github.com/gruzdev-dev/meddoc/app/models"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id string) (*models.Document, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Document, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, update models.DocumentUpdate) error
}
