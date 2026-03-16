package repositories

import (
	"context"
	"gin-quickstart/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShowtimeRepository struct {
	Collection *mongo.Collection
}

func NewShowtimeRepository(db *mongo.Database) *ShowtimeRepository {
	return &ShowtimeRepository{
		Collection: db.Collection("showtimes"),
	}
}

func (r *ShowtimeRepository) Create(ctx context.Context, item *models.Showtime) error {
	_, err := r.Collection.InsertOne(ctx, item)
	return err
}

func (r *ShowtimeRepository) GetAll(ctx context.Context) ([]models.Showtime, error) {
	var items []models.Showtime
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ShowtimeRepository) GetByID(ctx context.Context, id string) (*models.Showtime, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var item models.Showtime
	err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	return &item, err
}

func (r *ShowtimeRepository) Update(ctx context.Context, id string, updateData bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *ShowtimeRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *ShowtimeRepository) GetByMovieID(ctx context.Context, movieID string) ([]models.Showtime, error) {
	objID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, err
	}
	var items []models.Showtime
	cursor, err := r.Collection.Find(ctx, bson.M{"movie_id": objID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ShowtimeRepository) GetByDateRange(ctx context.Context, start time.Time, end time.Time) ([]models.Showtime, error) {
	var items []models.Showtime
	filter := bson.M{
		"start_time": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ShowtimeRepository) CheckOverlap(ctx context.Context, theaterNumber string, startTime, endTime time.Time, excludeID string) (bool, error) {
	filter := bson.M{
		"theater_number": theaterNumber,
		"start_time":     bson.M{"$lt": endTime},
		"end_time":       bson.M{"$gt": startTime},
	}

	if excludeID != "" && excludeID != "000000000000000000000000" {
		objID, err := primitive.ObjectIDFromHex(excludeID)
		if err == nil {
			filter["_id"] = bson.M{"$ne": objID}
		}
	}

	count, err := r.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
