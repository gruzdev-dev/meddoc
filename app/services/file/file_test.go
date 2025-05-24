package file

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/gruzdev-dev/meddoc/app/models"
)

const (
	testUserID    = "user123"
	testFileID    = "test-file-id.jpg"
	testContent   = "test content"
	largeFileSize = 2 << 20 // 2MB
)

type fakeMultipartFile struct {
	*bytes.Reader
}

func (f *fakeMultipartFile) Close() error {
	return nil
}

func (f *fakeMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Reader.ReadAt(p, off)
}

type testFile struct {
	content   []byte
	header    map[string][]string
	openError error
}

func (f *testFile) Open() (multipart.File, error) {
	if f.openError != nil {
		return nil, f.openError
	}
	return &fakeMultipartFile{bytes.NewReader(f.content)}, nil
}

func (f *testFile) GetFilename() string {
	return "test.jpg"
}

func (f *testFile) GetHeader() map[string][]string {
	return f.header
}

func (f *testFile) GetSize() int64 {
	return int64(len(f.content))
}

func createTestFile(content []byte) FileOpener {
	return &testFile{
		content: content,
		header: map[string][]string{
			"Content-Type": {"image/jpeg"},
		},
	}
}

func setupSuccessfulUploadMocks(mockStorage *MockStorage, mockRepo *MockFileRepository, fileID string, userID string, storageType string) {
	if storageType == "local" {
		mockStorage.EXPECT().
			Upload(gomock.Any(), fileID, gomock.Any()).
			Return(fileID, nil)
	} else {
		mockRepo.EXPECT().
			UploadFile(gomock.Any(), fileID[:len(fileID)-4], gomock.Any()).
			Return(fileID[:len(fileID)-4], nil)
	}

	mockRepo.EXPECT().
		Create(gomock.Any(), &models.File{
			ID:          fileID,
			UserID:      userID,
			DownloadURL: fmt.Sprintf("/files/%s", fileID),
			StorageType: storageType,
		}).
		Return(nil)
}

func TestService_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockStorage := NewMockStorage(ctrl)
	service := NewService(mockRepo, mockStorage)

	// Мокаем generateRandomName для предсказуемых ID
	originalGenerateRandomName := generateRandomName
	defer func() { generateRandomName = originalGenerateRandomName }()
	generateRandomName = func() (string, error) {
		return "test-file-id", nil
	}

	tests := []struct {
		name          string
		file          FileOpener
		userID        string
		setupMocks    func()
		expectedFile  *models.File
		expectedError string
	}{
		{
			name:   "successful_small_file_upload",
			file:   createTestFile([]byte(testContent)),
			userID: testUserID,
			setupMocks: func() {
				setupSuccessfulUploadMocks(mockStorage, mockRepo, testFileID, testUserID, "local")
			},
			expectedFile: &models.File{
				ID:          testFileID,
				UserID:      testUserID,
				DownloadURL: fmt.Sprintf("/files/%s", testFileID),
				StorageType: "local",
			},
		},
		{
			name:   "successful_large_file_upload",
			file:   createTestFile(make([]byte, largeFileSize)),
			userID: testUserID,
			setupMocks: func() {
				setupSuccessfulUploadMocks(mockStorage, mockRepo, testFileID, testUserID, "gridfs")
			},
			expectedFile: &models.File{
				ID:          testFileID,
				UserID:      testUserID,
				DownloadURL: fmt.Sprintf("/files/%s", testFileID),
				StorageType: "gridfs",
			},
		},
		{
			name:   "repository_error",
			file:   createTestFile([]byte(testContent)),
			userID: testUserID,
			setupMocks: func() {
				mockStorage.EXPECT().
					Upload(gomock.Any(), testFileID, gomock.Any()).
					Return(testFileID, nil)
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(stderrors.New("database error"))
			},
			expectedError: "database error",
		},
		{
			name:   "upload_error",
			file:   createTestFile([]byte(testContent)),
			userID: testUserID,
			setupMocks: func() {
				mockStorage.EXPECT().
					Upload(gomock.Any(), testFileID, gomock.Any()).
					Return("", stderrors.New("upload failed"))
			},
			expectedError: "upload failed",
		},
		{
			name: "file_open_error",
			file: &testFile{
				openError: stderrors.New("open failed"),
				header: map[string][]string{
					"Content-Type": {"image/jpeg"},
				},
			},
			userID:        testUserID,
			expectedError: "open failed",
		},
		{
			name:   "name_generation_error",
			file:   createTestFile([]byte(testContent)),
			userID: testUserID,
			setupMocks: func() {
				generateRandomName = func() (string, error) {
					return "", stderrors.New("generation failed")
				}
			},
			expectedError: "generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			file, err := service.UploadFile(context.Background(), tt.file, tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFile, file)
		})
	}
}

func TestService_DownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockStorage := NewMockStorage(ctrl)
	service := NewService(mockRepo, mockStorage)

	tests := []struct {
		name          string
		id            string
		userID        string
		setupMocks    func()
		expectedError string
		verifyContent bool
	}{
		{
			name:   "successful_local_file_download",
			id:     testFileID,
			userID: testUserID,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "test-file-id").
					Return(&models.File{
						ID:          testFileID,
						UserID:      testUserID,
						StorageType: "local",
					}, nil)
				mockStorage.EXPECT().
					Download(gomock.Any(), testFileID).
					Return(io.NopCloser(bytes.NewReader([]byte(testContent))), nil)
			},
			verifyContent: true,
		},
		{
			name:   "successful_gridfs_file_download",
			id:     testFileID,
			userID: testUserID,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "test-file-id").
					Return(&models.File{
						ID:          testFileID,
						UserID:      testUserID,
						StorageType: "gridfs",
					}, nil)
				mockRepo.EXPECT().
					DownloadFile(gomock.Any(), "test-file-id").
					Return(io.NopCloser(bytes.NewReader([]byte(testContent))), nil)
			},
			verifyContent: true,
		},
		{
			name:   "file_not_found",
			id:     testFileID,
			userID: testUserID,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "test-file-id").
					Return(nil, stderrors.New("not found"))
			},
			expectedError: "not found",
		},
		{
			name:   "access_denied",
			id:     testFileID,
			userID: testUserID,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "test-file-id").
					Return(&models.File{
						ID:          testFileID,
						UserID:      "other_user",
						StorageType: "local",
					}, nil)
			},
			expectedError: "access denied",
		},
		{
			name:   "unknown_storage_type",
			id:     testFileID,
			userID: testUserID,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "test-file-id").
					Return(&models.File{
						ID:          testFileID,
						UserID:      testUserID,
						StorageType: "unknown",
					}, nil)
			},
			expectedError: "unknown storage type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			reader, err := service.DownloadFile(context.Background(), tt.id, tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, reader)

			if tt.verifyContent {
				content, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, []byte(testContent), content)
			}
		})
	}
}

func TestService_UploadFile_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockStorage := NewMockStorage(ctrl)
	service := NewService(mockRepo, mockStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	file := createTestFile([]byte(testContent))
	mockStorage.EXPECT().
		Upload(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", context.DeadlineExceeded)

	_, err := service.UploadFile(ctx, file, testUserID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestService_DownloadFile_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileRepository(ctrl)
	mockStorage := NewMockStorage(ctrl)
	service := NewService(mockRepo, mockStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	mockRepo.EXPECT().
		GetByID(gomock.Any(), gomock.Any()).
		Return(nil, context.DeadlineExceeded)

	_, err := service.DownloadFile(ctx, testFileID, testUserID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
