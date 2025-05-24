//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gruzdev-dev/meddoc/app/models"
)

func TestAuthFlow(t *testing.T) {
	server, userService := setupTestServer(t)
	defer server.Close()

	regData := models.UserRegistration{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	t.Run("registration", func(t *testing.T) {
		body, err := json.Marshal(regData)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)
		assert.Equal(t, regData.Email, user.Email)
		assert.Equal(t, regData.Name, user.Name)
		assert.Empty(t, user.Password)
	})

	t.Run("login", func(t *testing.T) {
		loginData := models.UserLogin{
			Email:    regData.Email,
			Password: regData.Password,
		}

		body, err := json.Marshal(loginData)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokens models.TokenPair
		err = json.NewDecoder(resp.Body).Decode(&tokens)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		assert.Greater(t, tokens.ExpiresIn, 0)

		userID, err := userService.ValidateToken(tokens.AccessToken)
		require.NoError(t, err)
		assert.NotEmpty(t, userID)

		refreshData := models.RefreshToken{
			RefreshToken: tokens.RefreshToken,
		}

		body, err = json.Marshal(refreshData)
		require.NoError(t, err)

		resp, err = http.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var newTokens models.TokenPair
		err = json.NewDecoder(resp.Body).Decode(&newTokens)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.Greater(t, newTokens.ExpiresIn, 0)
	})

	t.Run("invalid login", func(t *testing.T) {
		loginData := models.UserLogin{
			Email:    regData.Email,
			Password: "wrongpassword",
		}

		body, err := json.Marshal(loginData)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshData := models.RefreshToken{
			RefreshToken: "invalid-token",
		}

		body, err := json.Marshal(refreshData)
		require.NoError(t, err)

		resp, err := http.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
