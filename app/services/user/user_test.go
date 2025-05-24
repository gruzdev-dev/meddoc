package user

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	cfg := Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}
	service := NewUserService(mockRepo, cfg)

	tests := []struct {
		name          string
		registration  models.UserRegistration
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful registration",
			registration: models.UserRegistration{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user *models.User) error {
						assert.NotEmpty(t, user.Password)
						assert.NotEqual(t, "password123", user.Password)
						return nil
					})
			},
			expectedError: nil,
		},
		{
			name: "user already exists",
			registration: models.UserRegistration{
				Email:    "existing@example.com",
				Name:     "Existing User",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.ErrUserExists)
			},
			expectedError: errors.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := service.Register(context.Background(), tt.registration)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.registration.Email, user.Email)
				assert.Equal(t, tt.registration.Name, user.Name)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	cfg := Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}
	service := NewUserService(mockRepo, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Password: string(hashedPassword),
	}

	tests := []struct {
		name          string
		email         string
		password      string
		mockSetup     func()
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "correct-password",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(existingUser, nil)
			},
			expectedError: nil,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrong-password",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(existingUser, nil)
			},
			expectedError: errors.ErrInvalidCredentials,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "any-password",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByEmail(gomock.Any(), "nonexistent@example.com").
					Return(nil, errors.ErrUserNotFound)
			},
			expectedError: errors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			tokens, err := service.Login(context.Background(), tt.email, tt.password)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Equal(t, int(cfg.AccessTokenTTL.Seconds()), tokens.ExpiresIn)
			}
		})
	}
}

func TestUserService_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockUserRepository(ctrl)
	cfg := Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}
	service := NewUserService(mockRepo, cfg)

	user := &models.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(cfg.RefreshTokenTTL).Unix(),
	})
	validRefreshToken, _ := refreshToken.SignedString([]byte(cfg.JWTSecret))

	tests := []struct {
		name          string
		refreshToken  string
		mockSetup     func()
		expectedError error
	}{
		{
			name:         "successful refresh",
			refreshToken: validRefreshToken,
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), user.ID).
					Return(user, nil)
			},
			expectedError: nil,
		},
		{
			name:          "invalid token",
			refreshToken:  "invalid-token",
			mockSetup:     func() {},
			expectedError: errors.ErrInvalidRefreshToken,
		},
		{
			name:         "user not found",
			refreshToken: validRefreshToken,
			mockSetup: func() {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), user.ID).
					Return(nil, errors.ErrUserNotFound)
			},
			expectedError: errors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			tokens, err := service.RefreshToken(context.Background(), tt.refreshToken)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokens)
				assert.NotEmpty(t, tokens.AccessToken)
				assert.NotEmpty(t, tokens.RefreshToken)
				assert.Equal(t, int(cfg.AccessTokenTTL.Seconds()), tokens.ExpiresIn)
			}
		})
	}
}

func TestUserService_ValidateToken(t *testing.T) {
	cfg := Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24,
	}
	service := NewUserService(nil, cfg)

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(cfg.AccessTokenTTL).Unix(),
	})
	validTokenString, _ := validToken.SignedString([]byte(cfg.JWTSecret))

	tests := []struct {
		name          string
		token         string
		expectedID    string
		expectedError error
	}{
		{
			name:          "valid token",
			token:         validTokenString,
			expectedID:    "user-123",
			expectedError: nil,
		},
		{
			name:          "invalid token",
			token:         "invalid-token",
			expectedID:    "",
			expectedError: errors.ErrInvalidToken,
		},
		{
			name:          "expired token",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsImV4cCI6MTUxNjIzOTAyMn0.2hDgYvYRtr7VZmHl2XGpGxJQzJQzJQzJQzJQzJQzJQ",
			expectedID:    "",
			expectedError: errors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := service.ValidateToken(tt.token)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
