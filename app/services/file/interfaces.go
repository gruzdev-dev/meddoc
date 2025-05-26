package file

import (
	"context"
	"io"

	"github.com/gruzdev-dev/meddoc/app/models"
)

type Storage interface {
	Upload(ctx context.Context, id string, reader io.Reader) error
	Download(ctx context.Context, id string) (io.ReadCloser, error)
}

type FileRepository interface {
	Create(ctx context.Context, file *models.FileCreation) (*models.FileRecord, error)
	GetByID(ctx context.Context, id string) (*models.FileRecord, error)
}
