package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
)

const (
	smallFileThreshold = 1 << 20 // 1MB
)

type Service struct {
	repo         FileRepository
	localStorage Storage
	gridStorage  Storage
}

func NewService(repo FileRepository, localStorage, gridStorage Storage) *Service {
	return &Service{
		repo:         repo,
		localStorage: localStorage,
		gridStorage:  gridStorage,
	}
}

func (s *Service) UploadFile(ctx context.Context, reader io.Reader, metadata models.FileMetadata, userID string) (*models.FileResponse, error) {
	storageType := "local"
	if metadata.Size >= smallFileThreshold {
		storageType = "gridfs"
	}

	fileCreation := &models.FileCreation{
		UserID:      userID,
		StorageType: storageType,
	}

	fileRecord, err := s.repo.Create(ctx, fileCreation)
	if err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	var err2 error
	if storageType == "local" {
		err2 = s.localStorage.Upload(ctx, fileRecord.ID, reader)
	} else {
		err2 = s.gridStorage.Upload(ctx, fileRecord.ID, reader)
	}

	if err2 != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err2)
	}

	return &models.FileResponse{
		ID: fileRecord.ID,
	}, nil
}

func (s *Service) DownloadFile(ctx context.Context, id string, userID string) (io.ReadCloser, error) {
	gridFSID := id
	if ext := filepath.Ext(id); ext != "" {
		gridFSID = id[:len(id)-len(ext)]
	}

	file, err := s.repo.GetByID(ctx, gridFSID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if file.UserID != userID {
		return nil, apperrors.ErrAccessDenied
	}

	var reader io.ReadCloser
	switch file.StorageType {
	case "gridfs":
		reader, err = s.gridStorage.Download(ctx, id)
	case "local":
		reader, err = s.localStorage.Download(ctx, id)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", file.StorageType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return reader, nil
}
