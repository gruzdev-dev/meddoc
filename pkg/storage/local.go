package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gruzdev-dev/meddoc/pkg/logger"
)

type Local struct {
	basePath string
}

func NewLocal(basePath string) (*Local, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &Local{basePath: basePath}, nil
}

func (s *Local) Upload(ctx context.Context, id string, reader io.Reader) error {
	filePath := filepath.Join(s.basePath, id)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error("failed to close file", err)
		}
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (s *Local) Download(ctx context.Context, id string) (io.ReadCloser, error) {
	filePath := filepath.Join(s.basePath, id)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}
