package file

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
	"github.com/stretchr/testify/assert"
)

func TestService_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockLocalStorage := NewMockStorage(ctrl)
	mockGridStorage := NewMockStorage(ctrl)

	service := NewService(mockRepo, mockLocalStorage, mockGridStorage)

	tests := []struct {
		name           string
		fileSize       int64
		expectedError  error
		setupMocks     func()
		expectedResult *models.FileResponse
	}{
		{
			name:     "successful small file upload",
			fileSize: 500 * 1024, // 500KB
			setupMocks: func() {
				fileCreation := &models.FileCreation{
					UserID:      "user123",
					StorageType: "local",
				}
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().Create(gomock.Any(), fileCreation).Return(fileRecord, nil)
				mockLocalStorage.EXPECT().Upload(gomock.Any(), "file123", gomock.Any()).Return(nil)
			},
			expectedResult: &models.FileResponse{
				ID: "file123",
			},
		},
		{
			name:     "successful large file upload",
			fileSize: 2 * 1024 * 1024, // 2MB
			setupMocks: func() {
				fileCreation := &models.FileCreation{
					UserID:      "user123",
					StorageType: "gridfs",
				}
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "gridfs",
				}
				mockRepo.EXPECT().Create(gomock.Any(), fileCreation).Return(fileRecord, nil)
				mockGridStorage.EXPECT().Upload(gomock.Any(), "file123", gomock.Any()).Return(nil)
			},
			expectedResult: &models.FileResponse{
				ID: "file123",
			},
		},
		{
			name:     "repository error",
			fileSize: 500 * 1024,
			setupMocks: func() {
				fileCreation := &models.FileCreation{
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().Create(gomock.Any(), fileCreation).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("failed to create file record: db error"),
		},
		{
			name:     "storage error",
			fileSize: 500 * 1024,
			setupMocks: func() {
				fileCreation := &models.FileCreation{
					UserID:      "user123",
					StorageType: "local",
				}
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().Create(gomock.Any(), fileCreation).Return(fileRecord, nil)
				mockLocalStorage.EXPECT().Upload(gomock.Any(), "file123", gomock.Any()).Return(errors.New("storage error"))
			},
			expectedError: errors.New("failed to upload file: storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			metadata := models.FileMetadata{
				Size: tt.fileSize,
			}
			reader := strings.NewReader("test content")

			result, err := service.UploadFile(context.Background(), reader, metadata, "user123")

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestService_DownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockLocalStorage := NewMockStorage(ctrl)
	mockGridStorage := NewMockStorage(ctrl)

	service := NewService(mockRepo, mockLocalStorage, mockGridStorage)

	tests := []struct {
		name          string
		fileID        string
		userID        string
		expectedError error
		setupMocks    func()
		expectedData  string
	}{
		{
			name:   "successful local file download",
			fileID: "file123",
			userID: "user123",
			setupMocks: func() {
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().GetByID(gomock.Any(), "file123").Return(fileRecord, nil)
				mockLocalStorage.EXPECT().Download(gomock.Any(), "file123").Return(io.NopCloser(strings.NewReader("test content")), nil)
			},
			expectedData: "test content",
		},
		{
			name:   "successful gridfs file download",
			fileID: "file123",
			userID: "user123",
			setupMocks: func() {
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "gridfs",
				}
				mockRepo.EXPECT().GetByID(gomock.Any(), "file123").Return(fileRecord, nil)
				mockGridStorage.EXPECT().Download(gomock.Any(), "file123").Return(io.NopCloser(strings.NewReader("test content")), nil)
			},
			expectedData: "test content",
		},
		{
			name:   "file not found",
			fileID: "file123",
			userID: "user123",
			setupMocks: func() {
				mockRepo.EXPECT().GetByID(gomock.Any(), "file123").Return(nil, apperrors.ErrNotFound)
			},
			expectedError: apperrors.ErrNotFound,
		},
		{
			name:   "access denied",
			fileID: "file123",
			userID: "user456",
			setupMocks: func() {
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().GetByID(gomock.Any(), "file123").Return(fileRecord, nil)
			},
			expectedError: apperrors.ErrAccessDenied,
		},
		{
			name:   "storage error",
			fileID: "file123",
			userID: "user123",
			setupMocks: func() {
				fileRecord := &models.FileRecord{
					ID:          "file123",
					UserID:      "user123",
					StorageType: "local",
				}
				mockRepo.EXPECT().GetByID(gomock.Any(), "file123").Return(fileRecord, nil)
				mockLocalStorage.EXPECT().Download(gomock.Any(), "file123").Return(nil, errors.New("storage error"))
			},
			expectedError: errors.New("failed to download file: storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			reader, err := service.DownloadFile(context.Background(), tt.fileID, tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrAccessDenied) {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
				assert.Nil(t, reader)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reader)

				data, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, string(data))

				err = reader.Close()
				assert.NoError(t, err)
			}
		})
	}
}
