package repository

import (
	"context"
	"time"

	"github.com/gruzdev-dev/meddoc/internal/domain"
	"github.com/gruzdev-dev/meddoc/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocumentRepository struct {
	collection *mongo.Collection
}

func NewDocumentRepository(db *mongo.Database) domain.DocumentRepository {
	collection := db.Collection("documents")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "title", Value: 1}},
			Options: options.Index().SetName("title_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_idx"),
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		logger.Error("failed to create indexes", err)
	}

	return &DocumentRepository{
		collection: collection,
	}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	// Set timestamps
	now := time.Now()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		logger.Error("failed to insert document into MongoDB", err)
		return domain.ErrDatabaseError
	}
	doc.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Document, error) {
	var doc domain.Document
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("document not found in MongoDB", "id", id)
			return nil, domain.ErrDocumentNotFound
		}
		logger.Error("failed to find document in MongoDB", err, "id", id)
		return nil, domain.ErrDatabaseError
	}
	return &doc, nil
}

func (r *DocumentRepository) GetAll(ctx context.Context) ([]*domain.Document, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		logger.Error("failed to find documents in MongoDB", err)
		return nil, domain.ErrDatabaseError
	}
	defer cursor.Close(ctx)

	var docs []*domain.Document
	if err = cursor.All(ctx, &docs); err != nil {
		logger.Error("failed to decode documents from MongoDB", err)
		return nil, domain.ErrDatabaseError
	}
	return docs, nil
}

func (r *DocumentRepository) Update(ctx context.Context, doc *domain.Document) error {
	// Update timestamp
	doc.UpdatedAt = time.Now()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": doc.ID}, doc)
	if err != nil {
		logger.Error("failed to update document in MongoDB", err, "id", doc.ID)
		return domain.ErrDatabaseError
	}
	if result.MatchedCount == 0 {
		logger.Warn("document not found for update in MongoDB", "id", doc.ID)
		return domain.ErrDocumentNotFound
	}
	return nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		logger.Error("failed to delete document from MongoDB", err, "id", id)
		return domain.ErrDatabaseError
	}
	if result.DeletedCount == 0 {
		logger.Warn("document not found for deletion in MongoDB", "id", id)
		return domain.ErrDocumentNotFound
	}
	return nil
}
