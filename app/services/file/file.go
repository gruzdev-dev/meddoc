package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	smallFileThreshold = 1 << 20 // 1MB
)

type Service struct {
	repo         FileRepository
	localStorage Storage
}

func NewService(repo FileRepository, localStorage Storage) *Service {
	return &Service{
		repo:         repo,
		localStorage: localStorage,
	}
}

func generateRandomName() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func determineFileExtension(filename, contentType string) string {
	if ext := filepath.Ext(filename); ext != "" {
		return ext
	}

	switch contentType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	default:
		return ""
	}
}

func (s *Service) uploadSmallFile(ctx context.Context, generatedName, ext string, src io.Reader) (string, error) {
	fileID := generatedName + ext
	_, err := s.localStorage.Upload(ctx, fileID, src)
	return fileID, err
}

func (s *Service) uploadLargeFile(generatedName, ext string, src io.Reader) (string, error) {
	fileID, err := s.repo.UploadFile(context.Background(), generatedName, src)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	return fileID + ext, nil
}

func (s *Service) UploadFile(ctx context.Context, file *multipart.FileHeader, userID string) (*models.File, error) {
	generatedName, err := generateRandomName()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random name: %w", err)
	}

	ext := determineFileExtension(file.Filename, file.Header.Get("Content-Type"))

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			logger.Error("failed to close source file", err)
		}
	}()

	var fileID string
	var storageType string
	if file.Size < smallFileThreshold {
		fileID, err = s.uploadSmallFile(ctx, generatedName, ext, src)
		storageType = "local"
	} else {
		fileID, err = s.uploadLargeFile(generatedName, ext, src)
		storageType = "gridfs"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	newFile := &models.File{
		ID:          fileID,
		UserID:      userID,
		DownloadURL: fmt.Sprintf("/files/%s", fileID),
		StorageType: storageType,
	}

	if err := s.repo.Create(ctx, newFile); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return newFile, nil
}

func (s *Service) DownloadFile(ctx context.Context, id string, userID string) (io.ReadCloser, error) {
	gridFSID := id
	if ext := filepath.Ext(id); ext != "" {
		gridFSID = id[:len(id)-len(ext)]
	}

	file, err := s.repo.GetByID(ctx, gridFSID)
	if err != nil {
		if stderrors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if file.UserID != userID {
		return nil, apperrors.ErrAccessDenied
	}

	switch file.StorageType {
	case "gridfs":
		return s.repo.DownloadFile(ctx, gridFSID)
	case "local":
		return s.localStorage.Download(ctx, id)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", file.StorageType)
	}
}
