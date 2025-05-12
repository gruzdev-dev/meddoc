package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gruzdev-dev/meddoc/internal/app/handler"
	"github.com/gruzdev-dev/meddoc/internal/app/repository"
	appservice "github.com/gruzdev-dev/meddoc/internal/app/service"
	"github.com/gruzdev-dev/meddoc/internal/database"
	"github.com/gruzdev-dev/meddoc/pkg/config"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"github.com/gruzdev-dev/meddoc/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load(filepath.Join("configs", "config.yaml"))
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger.Setup(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	// Initialize MongoDB connection
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

	// Initialize repositories
	docRepo := repository.NewDocumentRepository(mongoDB.Database())
	userRepo := repository.NewUserRepository(mongoDB.Database().Collection("users"))

	// Initialize services
	docService := appservice.NewDocumentService(docRepo)
	userService := appservice.NewUserService(userRepo, cfg.JWT.Secret)

	// Initialize handlers
	docHandler := handler.NewDocumentHandler(docService)
	userHandler := handler.NewUserHandler(userService)

	// Initialize router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.GinMiddleware(middleware.RequestID()))
	router.Use(middleware.GinMiddleware(middleware.Logging()))
	router.Use(middleware.GinMiddleware(middleware.Recovery()))
	router.Use(middleware.GinMiddleware(middleware.Compression()))
	router.Use(middleware.GinMiddleware(middleware.SecurityHeaders()))

	// Initialize routes
	api := router.Group("/api/v1")
	{
		docHandler.RegisterRoutes(api)
		userHandler.RegisterRoutes(api)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		logger.Fatal("server error", err)

	case sig := <-shutdown:
		logger.Info("shutting down server...", "signal", sig)

		// Give outstanding requests 15 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("server forced to shutdown", err)
		}
	}

	logger.Info("server exited")
}
