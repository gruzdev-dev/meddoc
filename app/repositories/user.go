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
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository struct {
	collection *mongo.Collection
}

type mongoUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Name      string             `bson:"name"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: collection,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	var existingUser mongoUser
	err := r.collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return ErrUserExists
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	mongoUser := mongoUser{
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	result, err := r.collection.InsertOne(ctx, mongoUser)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var mongoUser mongoUser
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&mongoUser)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        mongoUser.ID.Hex(),
		Email:     mongoUser.Email,
		Name:      mongoUser.Name,
		Password:  mongoUser.Password,
		CreatedAt: mongoUser.CreatedAt,
		UpdatedAt: mongoUser.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mongoUser mongoUser
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoUser)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        mongoUser.ID.Hex(),
		Email:     mongoUser.Email,
		Name:      mongoUser.Name,
		Password:  mongoUser.Password,
		CreatedAt: mongoUser.CreatedAt,
		UpdatedAt: mongoUser.UpdatedAt,
	}, nil
}
