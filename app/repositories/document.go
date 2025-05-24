package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/gruzdev-dev/meddoc/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
)

type DocumentRepository struct {
	collection *mongo.Collection
}

type mongoDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Description string             `bson:"description,omitempty"`
	Date        string             `bson:"date,omitempty"`
	File        string             `bson:"file,omitempty"`
	Category    string             `bson:"category,omitempty"`
	Priority    int                `bson:"priority,omitempty"`
	Content     map[string]string  `bson:"content,omitempty"`
	UserID      string             `bson:"user_id"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func NewDocumentRepository(collection *mongo.Collection) *DocumentRepository {
	return &DocumentRepository{
		collection: collection,
	}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	mongoDoc := mongoDocument{
		Title:       doc.Title,
		Description: doc.Description,
		Date:        doc.Date,
		File:        doc.File,
		Category:    doc.Category,
		Priority:    doc.Priority,
		Content:     doc.Content,
		UserID:      doc.UserID,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}

	result, err := r.collection.InsertOne(ctx, mongoDoc)
	if err != nil {
		return err
	}

	doc.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoDoc mongoDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoDoc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrDocumentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &models.Document{
		ID:          mongoDoc.ID.Hex(),
		Title:       mongoDoc.Title,
		Description: mongoDoc.Description,
		Date:        mongoDoc.Date,
		File:        mongoDoc.File,
		Category:    mongoDoc.Category,
		Priority:    mongoDoc.Priority,
		Content:     mongoDoc.Content,
		UserID:      mongoDoc.UserID,
		CreatedAt:   mongoDoc.CreatedAt,
		UpdatedAt:   mongoDoc.UpdatedAt,
	}, nil
}

func (r *DocumentRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []*models.Document
	for cursor.Next(ctx) {
		var mongoDoc mongoDocument
		if err := cursor.Decode(&mongoDoc); err != nil {
			return nil, err
		}

		doc := &models.Document{
			ID:          mongoDoc.ID.Hex(),
			Title:       mongoDoc.Title,
			Description: mongoDoc.Description,
			Date:        mongoDoc.Date,
			File:        mongoDoc.File,
			Category:    mongoDoc.Category,
			Priority:    mongoDoc.Priority,
			Content:     mongoDoc.Content,
			UserID:      mongoDoc.UserID,
			CreatedAt:   mongoDoc.CreatedAt,
			UpdatedAt:   mongoDoc.UpdatedAt,
		}
		documents = append(documents, doc)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (r *DocumentRepository) Update(ctx context.Context, id string, update models.DocumentUpdate) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	set := bson.M{}
	if update.Title != nil {
		set["title"] = *update.Title
	}
	if update.Description != nil {
		set["description"] = *update.Description
	}
	if update.Date != nil {
		set["date"] = *update.Date
	}
	if update.File != nil {
		set["file"] = *update.File
	}
	if update.Category != nil {
		set["category"] = *update.Category
	}
	if update.Priority != nil {
		set["priority"] = *update.Priority
	}
	if update.Content != nil {
		set["content"] = update.Content
	}
	set["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": set},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}
	return nil
}
