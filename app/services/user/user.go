package user

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gruzdev-dev/meddoc/app/config"
	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
)

type Config struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type UserService struct {
	repo            UserRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewUserService(repo UserRepository, cfg Config) *UserService {
	return &UserService{
		repo:            repo,
		jwtSecret:       []byte(cfg.JWTSecret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func NewUserServiceFromConfig(repo UserRepository, cfg *config.Config) *UserService {
	return NewUserService(repo, Config{
		JWTSecret:       cfg.Auth.Secret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	})
}

func (s *UserService) Register(ctx context.Context, reg models.UserRegistration) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    reg.Email,
		Name:     reg.Name,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, apperrors.ErrUserExists) {
			return nil, apperrors.ErrUserExists
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*models.TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, apperrors.ErrInvalidRefreshToken
	}

	user, err := s.repo.GetByID(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

func (s *UserService) generateTokenPair(user *models.User) (*models.TokenPair, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.accessTokenTTL).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.refreshTokenTTL).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *UserService) ValidateToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", apperrors.ErrInvalidToken
	}

	return claims.Subject, nil
}
