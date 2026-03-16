package repositories

import (
	"context"
	"gin-quickstart/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MovieRepository struct {
	Collection *mongo.Collection
}

func NewMovieRepository(db *mongo.Database) *MovieRepository {
	return &MovieRepository{
		Collection: db.Collection("movies"),
	}
}

func (r *MovieRepository) Create(ctx context.Context, item *models.Movie) error {
	_, err := r.Collection.InsertOne(ctx, item)
	return err
}

func (r *MovieRepository) GetAll(ctx context.Context) ([]models.Movie, error) {
	var items []models.Movie
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MovieRepository) GetByID(ctx context.Context, id string) (*models.Movie, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var item models.Movie
	err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	return &item, err
}

func (r *MovieRepository) GetByTitle(ctx context.Context, title string) (*models.Movie, error) {
	var item models.Movie
	err := r.Collection.FindOne(ctx, bson.M{"title": title}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *MovieRepository) Update(ctx context.Context, id string, updateData bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *MovieRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
