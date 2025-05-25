package repositories

import (
	"context"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"

	"github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
)

type FileRepository struct {
	collection *mongo.Collection
	bucket     *gridfs.Bucket
}

func NewFileRepository(collection *mongo.Collection) (*FileRepository, error) {
	bucket, err := gridfs.NewBucket(collection.Database())
	if err != nil {
		return nil, fmt.Errorf("failed to create GridFS bucket: %w", err)
	}

	return &FileRepository{
		collection: collection,
		bucket:     bucket,
	}, nil
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	mongoFile := bson.M{
		"user_id":      file.UserID,
		"download_url": file.DownloadURL,
		"storage_type": file.StorageType,
	}

	result, err := r.collection.InsertOne(ctx, mongoFile)
	if err != nil {
		return err
	}

	file.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*models.File, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoFile bson.M
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoFile)
	if err != nil {
		return nil, err
	}

	storageType, ok := mongoFile["storage_type"].(string)
	if !ok {
		storageType = ""
	}

	return &models.File{
		ID:          mongoFile["_id"].(primitive.ObjectID).Hex(),
		UserID:      mongoFile["user_id"].(string),
		DownloadURL: mongoFile["download_url"].(string),
		StorageType: storageType,
	}, nil
}

func (r *FileRepository) UploadFile(ctx context.Context, filename string, reader io.Reader) (string, error) {
	uploadStream, err := r.bucket.OpenUploadStream(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := uploadStream.Close(); err != nil {
			logger.Error("failed to close upload stream", err)
		}
	}()

	if _, err = io.Copy(uploadStream, reader); err != nil {
		return "", err
	}

	return uploadStream.FileID.(primitive.ObjectID).Hex(), nil
}

func (r *FileRepository) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	stream, err := r.bucket.OpenDownloadStream(objectID)
	if err != nil {
		if err == gridfs.ErrFileNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to open download stream: %w", err)
	}

	return stream, nil
}
