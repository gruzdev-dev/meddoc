package database

import (
	"context"
	"fmt"
	"time"

	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoConnectTimeout = 10 * time.Second
	mongoRetryInterval  = 5 * time.Second
	mongoMaxRetries     = 3
)

// MongoDBConfig содержит конфигурацию для подключения к MongoDB
type MongoDBConfig struct {
	URI      string
	Database string
}

// MongoDB представляет соединение с MongoDB
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   MongoDBConfig
}

// NewMongoDB создает новое соединение с MongoDB
func NewMongoDB(ctx context.Context, cfg MongoDBConfig) (*MongoDB, error) {
	var client *mongo.Client
	var err error

	for i := 0; i < mongoMaxRetries; i++ {
		logger.Info("connecting to MongoDB", "attempt", i+1, "uri", cfg.URI)
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
		if err == nil {
			// Ping the database to verify connection
			if err = client.Ping(ctx, nil); err == nil {
				logger.Info("successfully connected to MongoDB")
				return &MongoDB{
					client:   client,
					database: client.Database(cfg.Database),
					config:   cfg,
				}, nil
			}
		}
		logger.Error("failed to connect to MongoDB", err)
		if i < mongoMaxRetries-1 {
			time.Sleep(mongoRetryInterval)
		}
	}

	return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", mongoMaxRetries, err)
}

// Close закрывает соединение с MongoDB
func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		logger.Error("failed to disconnect from MongoDB", err)
		return err
	}
	return nil
}

// Database возвращает базу данных
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Client возвращает клиент MongoDB
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}
