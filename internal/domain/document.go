package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DocumentType represents the type of medical document
type DocumentType string

const (
	DocumentTypePrescription DocumentType = "prescription"
	DocumentTypeDiagnosis    DocumentType = "diagnosis"
	DocumentTypeLabResult    DocumentType = "lab_result"
)

// Document represents a medical document
type Document struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// DocumentRepository defines the interface for document storage operations
type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Document, error)
	GetAll(ctx context.Context) ([]*Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
