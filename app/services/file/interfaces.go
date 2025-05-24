package file

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/gruzdev-dev/meddoc/app/models"
)

type FileOpener interface {
	Open() (multipart.File, error)
	GetFilename() string
	GetHeader() map[string][]string
	GetSize() int64
}

type Storage interface {
	Upload(ctx context.Context, filename string, reader io.Reader) (string, error)
	Download(ctx context.Context, fileID string) (io.ReadCloser, error)
}

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id string) (*models.File, error)
	UploadFile(ctx context.Context, filename string, reader io.Reader) (string, error)
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)
}
