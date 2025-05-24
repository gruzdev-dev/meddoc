package document

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
)

func TestService_CreateDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDocumentRepository(ctrl)
	service := NewService(mockRepo)

	tests := []struct {
		name          string
		creation      models.DocumentCreation
		userID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful creation",
			creation: models.DocumentCreation{
				Title:       "Test Document",
				Description: "Test Description",
				Date:        "2024-03-20",
				File:        "test.pdf",
				Category:    "test",
				Priority:    1,
				Content:     map[string]string{"key": "value"},
			},
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, doc *models.Document) error {
						assert.Equal(t, "Test Document", doc.Title)
						assert.Equal(t, "Test Description", doc.Description)
						assert.Equal(t, "2024-03-20", doc.Date)
						assert.Equal(t, "test.pdf", doc.File)
						assert.Equal(t, "test", doc.Category)
						assert.Equal(t, 1, doc.Priority)
						assert.Equal(t, map[string]string{"key": "value"}, doc.Content)
						assert.Equal(t, "user-123", doc.UserID)
						return nil
					})
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			creation: models.DocumentCreation{
				Title: "Test Document",
			},
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.ErrInternal)
			},
			expectedError: errors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			doc, err := service.CreateDocument(context.Background(), tt.creation, tt.userID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, tt.creation.Title, doc.Title)
				assert.Equal(t, tt.creation.Description, doc.Description)
				assert.Equal(t, tt.creation.Date, doc.Date)
				assert.Equal(t, tt.creation.File, doc.File)
				assert.Equal(t, tt.creation.Category, doc.Category)
				assert.Equal(t, tt.creation.Priority, doc.Priority)
				assert.Equal(t, tt.creation.Content, doc.Content)
				assert.Equal(t, tt.userID, doc.UserID)
			}
		})
	}
}

func TestService_GetDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDocumentRepository(ctrl)
	service := NewService(mockRepo)

	existingDoc := &models.Document{
		ID:          "doc-123",
		Title:       "Test Document",
		Description: "Test Description",
		UserID:      "user-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		docID         string
		userID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "successful get",
			docID:  "doc-123",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
			},
			expectedError: nil,
		},
		{
			name:   "document not found",
			docID:  "nonexistent",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, errors.ErrDocumentNotFound)
			},
			expectedError: errors.ErrDocumentNotFound,
		},
		{
			name:   "access denied",
			docID:  "doc-123",
			userID: "other-user",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
			},
			expectedError: errors.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			doc, err := service.GetDocument(context.Background(), tt.docID, tt.userID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, existingDoc.ID, doc.ID)
				assert.Equal(t, existingDoc.Title, doc.Title)
				assert.Equal(t, existingDoc.Description, doc.Description)
				assert.Equal(t, existingDoc.UserID, doc.UserID)
			}
		})
	}
}

func TestService_GetUserDocuments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDocumentRepository(ctrl)
	service := NewService(mockRepo)

	userDocs := []*models.Document{
		{
			ID:          "doc-1",
			Title:       "Document 1",
			Description: "Description 1",
			UserID:      "user-123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "doc-2",
			Title:       "Document 2",
			Description: "Description 2",
			UserID:      "user-123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name          string
		userID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "successful get",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByUserID(gomock.Any(), "user-123").
					Return(userDocs, nil)
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByUserID(gomock.Any(), "user-123").
					Return(nil, errors.ErrInternal)
			},
			expectedError: errors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			docs, err := service.GetUserDocuments(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, docs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, docs)
				assert.Len(t, docs, 2)
				assert.Equal(t, userDocs[0].ID, docs[0].ID)
				assert.Equal(t, userDocs[0].Title, docs[0].Title)
				assert.Equal(t, userDocs[1].ID, docs[1].ID)
				assert.Equal(t, userDocs[1].Title, docs[1].Title)
			}
		})
	}
}

func TestService_DeleteDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDocumentRepository(ctrl)
	service := NewService(mockRepo)

	existingDoc := &models.Document{
		ID:     "doc-123",
		Title:  "Test Document",
		UserID: "user-123",
	}

	tests := []struct {
		name          string
		docID         string
		userID        string
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "successful delete",
			docID:  "doc-123",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
				mockRepo.EXPECT().
					Delete(gomock.Any(), "doc-123").
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "document not found",
			docID:  "nonexistent",
			userID: "user-123",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, errors.ErrDocumentNotFound)
			},
			expectedError: errors.ErrDocumentNotFound,
		},
		{
			name:   "access denied",
			docID:  "doc-123",
			userID: "other-user",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
			},
			expectedError: errors.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.DeleteDocument(context.Background(), tt.docID, tt.userID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_UpdateDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockDocumentRepository(ctrl)
	service := NewService(mockRepo)

	existingDoc := &models.Document{
		ID:          "doc-123",
		Title:       "Original Title",
		Description: "Original Description",
		UserID:      "user-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedDoc := &models.Document{
		ID:          "doc-123",
		Title:       "Updated Title",
		Description: "Updated Description",
		UserID:      "user-123",
		CreatedAt:   existingDoc.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		docID         string
		userID        string
		update        models.DocumentUpdate
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "successful update",
			docID:  "doc-123",
			userID: "user-123",
			update: models.DocumentUpdate{
				Title:       stringPtr("Updated Title"),
				Description: stringPtr("Updated Description"),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
				mockRepo.EXPECT().
					Update(gomock.Any(), "doc-123", gomock.Any()).
					Return(nil)
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(updatedDoc, nil)
			},
			expectedError: nil,
		},
		{
			name:   "document not found",
			docID:  "nonexistent",
			userID: "user-123",
			update: models.DocumentUpdate{
				Title: stringPtr("New Title"),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, errors.ErrDocumentNotFound)
			},
			expectedError: errors.ErrDocumentNotFound,
		},
		{
			name:   "access denied",
			docID:  "doc-123",
			userID: "other-user",
			update: models.DocumentUpdate{
				Title: stringPtr("New Title"),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), "doc-123").
					Return(existingDoc, nil)
			},
			expectedError: errors.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			doc, err := service.UpdateDocument(context.Background(), tt.docID, tt.update, tt.userID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, updatedDoc.ID, doc.ID)
				assert.Equal(t, updatedDoc.Title, doc.Title)
				assert.Equal(t, updatedDoc.Description, doc.Description)
				assert.Equal(t, updatedDoc.UserID, doc.UserID)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
