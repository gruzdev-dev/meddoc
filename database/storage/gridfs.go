package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
)

type GridFSStorage struct {
	bucket *gridfs.Bucket
}

func NewGridFSStorage(db *mongo.Database) (*GridFSStorage, error) {
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create GridFS bucket: %w", err)
	}

	return &GridFSStorage{
		bucket: bucket,
	}, nil
}

func (s *GridFSStorage) Upload(ctx context.Context, id string, reader io.Reader) error {
	uploadStream, err := s.bucket.OpenUploadStream(id)
	if err != nil {
		return fmt.Errorf("failed to open upload stream: %w", err)
	}
	defer func() {
		if err := uploadStream.Close(); err != nil {
			logger.Error("failed to close upload stream", err)
		}
	}()

	if _, err := io.Copy(uploadStream, reader); err != nil {
		return fmt.Errorf("failed to copy file to upload stream: %w", err)
	}

	return nil
}

func (s *GridFSStorage) Download(ctx context.Context, id string) (io.ReadCloser, error) {
	stream, err := s.bucket.OpenDownloadStream(id)
	if err != nil {
		if errors.Is(err, gridfs.ErrFileNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to open download stream: %w", err)
	}

	return stream, nil
}
