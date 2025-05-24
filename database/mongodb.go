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

type MongoDBConfig struct {
	URI      string
	Database string
}

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   MongoDBConfig
}

func NewMongoDB(ctx context.Context, cfg MongoDBConfig) (*MongoDB, error) {
	var client *mongo.Client
	var err error

	for i := 0; i < mongoMaxRetries; i++ {
		logger.Info("connecting to MongoDB", "attempt", i+1, "uri", cfg.URI)
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
		if err == nil {
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

func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		logger.Error("failed to disconnect from MongoDB", err)
		return err
	}
	return nil
}

func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

func (m *MongoDB) Client() *mongo.Client {
	return m.client
}
