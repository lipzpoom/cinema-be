package repositories

import (
	"context"
	"gin-quickstart/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository struct {
	Collection *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		Collection: db.Collection("bookings"),
	}
}

func (r *BookingRepository) Create(ctx context.Context, item *models.Booking) error {
	_, err := r.Collection.InsertOne(ctx, item)
	return err
}

func (r *BookingRepository) GetAll(ctx context.Context) ([]models.Booking, error) {
	var items []models.Booking
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var item models.Booking
	err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	return &item, err
}

func (r *BookingRepository) Update(ctx context.Context, id string, updateData bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *BookingRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID string) ([]models.Booking, error) {
	var items []models.Booking
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.Collection.Find(ctx, bson.M{"user_id": objID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BookingRepository) HasBookingForSeatNumber(ctx context.Context, showtimeID primitive.ObjectID, seatNumber string) (bool, error) {
	count, err := r.Collection.CountDocuments(ctx, bson.M{
		"showtime_id": showtimeID,
		"seats":       seatNumber,
		"status":      bson.M{"$ne": models.BookingFailed},
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
