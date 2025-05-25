//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gruzdev-dev/meddoc/app/models"
)

func TestFileFlow(t *testing.T) {
	server, _ := setupTestServer(t)
	defer server.Close()

	regData := models.UserRegistration{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	body, err := json.Marshal(regData)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loginData := models.UserLogin{
		Email:    regData.Email,
		Password: regData.Password,
	}

	body, err = json.Marshal(loginData)
	require.NoError(t, err)

	resp, err = http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokens models.TokenPair
	err = json.NewDecoder(resp.Body).Decode(&tokens)
	require.NoError(t, err)

	smallFileContent := []byte("This is a small test file")
	largeFileContent := make([]byte, 10<<20) // 10MB файл

	smallFilePath := filepath.Join(t.TempDir(), "small.jpg")
	largeFilePath := filepath.Join(t.TempDir(), "large.jpg")

	err = os.WriteFile(smallFilePath, smallFileContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(largeFilePath, largeFileContent, 0644)
	require.NoError(t, err)

	t.Run("upload small file", func(t *testing.T) {
		file, err := os.Open(smallFilePath)
		require.NoError(t, err)
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "small.jpg")
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/files/upload", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-File-Type", "image/jpeg")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var uploadedFile models.File
		err = json.NewDecoder(resp.Body).Decode(&uploadedFile)
		require.NoError(t, err)
		assert.NotEmpty(t, uploadedFile.ID)
		assert.NotEmpty(t, uploadedFile.DownloadURL)

		t.Run("download small file", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+uploadedFile.DownloadURL, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			downloadedContent, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, smallFileContent, downloadedContent)
		})
	})

	t.Run("upload large file", func(t *testing.T) {
		file, err := os.Open(largeFilePath)
		require.NoError(t, err)
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "large.jpg")
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/files/upload", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-File-Type", "image/jpeg")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var uploadedFile models.File
		err = json.NewDecoder(resp.Body).Decode(&uploadedFile)
		require.NoError(t, err)
		assert.NotEmpty(t, uploadedFile.ID)
		assert.NotEmpty(t, uploadedFile.DownloadURL)

		t.Run("download large file", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+uploadedFile.DownloadURL, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			downloadedContent, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, len(largeFileContent), len(downloadedContent))
		})
	})

	t.Run("upload invalid file type", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "invalid.*.exe")
		require.NoError(t, err)
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "invalid.exe")
		require.NoError(t, err)
		_, err = io.Copy(part, file)
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/files/upload", body)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-File-Type", "application/octet-stream")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("download non-existent file", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/files/507f1f77bcf86cd799439011", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
