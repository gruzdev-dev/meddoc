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

func TestDocumentFlow(t *testing.T) {
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

	t.Run("create document", func(t *testing.T) {
		docData := models.DocumentCreation{
			Title:       "Test Document",
			Description: "Test Description",
			Date:        "2024-03-20",
			Category:    "Test",
			Priority:    1,
			Content: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}

		body, err := json.Marshal(docData)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/documents", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var doc models.Document
		err = json.NewDecoder(resp.Body).Decode(&doc)
		require.NoError(t, err)
		assert.Equal(t, docData.Title, doc.Title)
		assert.Equal(t, docData.Description, doc.Description)
		assert.Equal(t, docData.Date, doc.Date)
		assert.Equal(t, docData.Category, doc.Category)
		assert.Equal(t, docData.Priority, doc.Priority)
		assert.Equal(t, docData.Content, doc.Content)
		assert.NotEmpty(t, doc.ID)

		t.Run("get document", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/documents/"+doc.ID, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var gotDoc models.Document
			err = json.NewDecoder(resp.Body).Decode(&gotDoc)
			require.NoError(t, err)
			assert.Equal(t, doc.ID, gotDoc.ID)
			assert.Equal(t, doc.Title, gotDoc.Title)
		})

		t.Run("update document", func(t *testing.T) {
			updateData := models.DocumentUpdate{
				Title:       stringPtr("Updated Title"),
				Description: stringPtr("Updated Description"),
				Priority:    intPtr(2),
			}

			body, err := json.Marshal(updateData)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPatch, server.URL+"/api/v1/documents/"+doc.ID, bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var updatedDoc models.Document
			err = json.NewDecoder(resp.Body).Decode(&updatedDoc)
			require.NoError(t, err)
			assert.Equal(t, *updateData.Title, updatedDoc.Title)
			assert.Equal(t, *updateData.Description, updatedDoc.Description)
			assert.Equal(t, *updateData.Priority, updatedDoc.Priority)
		})

		t.Run("delete document", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/documents/"+doc.ID, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			req, err = http.NewRequest(http.MethodGet, server.URL+"/api/v1/documents/"+doc.ID, nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
