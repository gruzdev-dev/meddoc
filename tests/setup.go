//go:build integration

package tests

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
	"github.com/gruzdev-dev/meddoc/app/handlers"
	"github.com/gruzdev-dev/meddoc/app/server/middleware"
	"github.com/gruzdev-dev/meddoc/app/services/document"
	"github.com/gruzdev-dev/meddoc/app/services/file"
	"github.com/gruzdev-dev/meddoc/app/services/user"
	"github.com/gruzdev-dev/meddoc/config"
	"github.com/gruzdev-dev/meddoc/database"
	"github.com/gruzdev-dev/meddoc/database/repositories"
	"github.com/gruzdev-dev/meddoc/pkg/storage"
)

func setupTestServer(t *testing.T) (*httptest.Server, *user.UserService) {
	cfg, err := config.Load("test_config.yaml")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoCfg := database.MongoDBConfig{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
	}
	mongoDB, err := database.NewMongoDB(ctx, mongoCfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoDB.Close(ctx)
	})

	err = mongoDB.Database().Drop(ctx)
	require.NoError(t, err)

	testStorageDir := "test_storage"
	require.NoError(t, os.MkdirAll(testStorageDir, 0755))
	t.Cleanup(func() {
		// Wait for all file operations to complete
		time.Sleep(100 * time.Millisecond)
		_ = os.RemoveAll(testStorageDir)
	})

	userRepo := repositories.NewUserRepository(mongoDB.Database().Collection("users"))
	userService := user.NewUserServiceFromConfig(userRepo, cfg)
	fileRepo, err := repositories.NewFileRepository(mongoDB.Database().Collection("files"))
	require.NoError(t, err)
	localStorage, err := storage.NewLocal(testStorageDir)
	require.NoError(t, err)
	documentRepo := repositories.NewDocumentRepository(mongoDB.Database().Collection("documents"))
	documentService := document.NewService(documentRepo)
	fileService := file.NewService(fileRepo, localStorage)

	handlers := handlers.NewHandlers(userService, documentService, fileService)
	router := mux.NewRouter()
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging())
	router.Use(middleware.Recovery())
	router.Use(middleware.Compression())
	router.Use(middleware.SecurityHeaders())
	api := router.PathPrefix("/api/v1").Subrouter()
	handlers.RegisterRoutes(api)

	return httptest.NewServer(router), userService
}
