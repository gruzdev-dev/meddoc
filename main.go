package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gruzdev-dev/meddoc/app/handlers"
	"github.com/gruzdev-dev/meddoc/app/server"
	"github.com/gruzdev-dev/meddoc/app/services/document"
	"github.com/gruzdev-dev/meddoc/app/services/file"
	"github.com/gruzdev-dev/meddoc/app/services/user"
	"github.com/gruzdev-dev/meddoc/config"
	"github.com/gruzdev-dev/meddoc/database"
	"github.com/gruzdev-dev/meddoc/database/repositories"
	dbstorage "github.com/gruzdev-dev/meddoc/database/storage"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	localstorage "github.com/gruzdev-dev/meddoc/pkg/storage"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Setup(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoCfg := database.MongoDBConfig{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
	}
	mongoDB, err := database.NewMongoDB(ctx, mongoCfg)
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoDB.Close(ctx); err != nil {
			logger.Error("failed to disconnect from MongoDB", err)
		}
	}()

	userRepo := repositories.NewUserRepository(mongoDB.Database().Collection("users"))
	userService := user.NewUserServiceFromConfig(userRepo, cfg)

	documentRepo := repositories.NewDocumentRepository(mongoDB.Database().Collection("documents"))
	documentService := document.NewService(documentRepo)

	fileRepo := repositories.NewFileRepository(mongoDB.Database().Collection("files"))

	localStorage, err := localstorage.NewLocal("storage/files")
	if err != nil {
		logger.Fatal("failed to create local storage", err)
	}

	gridStorage, err := dbstorage.NewGridFSStorage(mongoDB.Database())
	if err != nil {
		logger.Fatal("failed to create grid storage", err)
	}

	fileService := file.NewService(fileRepo, localStorage, gridStorage)

	handlers := handlers.NewHandlers(userService, documentService, fileService)

	srv := server.NewServer(cfg, handlers)
	if err := srv.Start(); err != nil {
		logger.Fatal("server error", err)
	}
}
