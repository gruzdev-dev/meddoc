package repositories

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	apperrors "github.com/gruzdev-dev/meddoc/app/errors"
	"github.com/gruzdev-dev/meddoc/app/models"
)

type mongoFileRecord struct {
	ID          primitive.ObjectID `bson:"_id"`
	UserID      string             `bson:"user_id"`
	StorageType string             `bson:"storage_type"`
}

func toMongoFileRecord(file *models.FileCreation) bson.M {
	return bson.M{
		"user_id":      file.UserID,
		"storage_type": file.StorageType,
	}
}

func fromMongoFileRecord(mongoFile mongoFileRecord) *models.FileRecord {
	return &models.FileRecord{
		ID:          mongoFile.ID.Hex(),
		UserID:      mongoFile.UserID,
		StorageType: mongoFile.StorageType,
	}
}

type FileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(collection *mongo.Collection) *FileRepository {
	return &FileRepository{
		collection: collection,
	}
}

func (r *FileRepository) Create(ctx context.Context, file *models.FileCreation) (*models.FileRecord, error) {
	mongoFile := toMongoFileRecord(file)

	result, err := r.collection.InsertOne(ctx, mongoFile)
	if err != nil {
		return nil, err
	}

	id := result.InsertedID.(primitive.ObjectID)
	return &models.FileRecord{
		ID:          id.Hex(),
		UserID:      file.UserID,
		StorageType: file.StorageType,
	}, nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*models.FileRecord, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoFile mongoFileRecord
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoFile)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	return fromMongoFileRecord(mongoFile), nil
}
