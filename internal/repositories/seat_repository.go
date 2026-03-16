package repositories

import (
	"context"
	"gin-quickstart/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SeatRepository struct {
	Collection *mongo.Collection
}

func NewSeatRepository(db *mongo.Database) *SeatRepository {
	return &SeatRepository{
		Collection: db.Collection("seats"),
	}
}

func (r *SeatRepository) Create(ctx context.Context, item *models.Seat) error {
	_, err := r.Collection.InsertOne(ctx, item)
	return err
}

func (r *SeatRepository) GetAll(ctx context.Context) ([]models.Seat, error) {
	var items []models.Seat
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SeatRepository) GetByID(ctx context.Context, id string) (*models.Seat, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var item models.Seat
	err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	return &item, err
}

func (r *SeatRepository) Update(ctx context.Context, id string, updateData bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *SeatRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *SeatRepository) LockAvailableSeats(ctx context.Context, showtimeID primitive.ObjectID, seatNumbers []string) error {
	filter := bson.M{
		"showtime_id": showtimeID,
		"seat_number": bson.M{"$in": seatNumbers},
		"status":      models.SeatAvailable,
	}
	update := bson.M{"$set": bson.M{"status": models.SeatLocked}}

	result, err := r.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// If the number of modified seats doesn't match the requested seats,
	// it means someone else locked them or they don't exist. Rollback needed.
	if result.ModifiedCount != int64(len(seatNumbers)) {
		if result.ModifiedCount > 0 {
			// Rollback the ones we successfully locked
			rollbackFilter := bson.M{
				"showtime_id": showtimeID,
				"seat_number": bson.M{"$in": seatNumbers},
				"status":      models.SeatLocked,
			}
			rollbackUpdate := bson.M{"$set": bson.M{"status": models.SeatAvailable}}
			_, _ = r.Collection.UpdateMany(ctx, rollbackFilter, rollbackUpdate)
		}
		return mongo.ErrNoDocuments // or return a custom error
	}

	return nil
}

func (r *SeatRepository) UpdateStatusByShowtimeAndSeats(ctx context.Context, showtimeID primitive.ObjectID, seatNumbers []string, status models.SeatStatus) error {
	filter := bson.M{
		"showtime_id": showtimeID,
		"seat_number": bson.M{"$in": seatNumbers},
	}
	update := bson.M{"$set": bson.M{"status": status}}
	_, err := r.Collection.UpdateMany(ctx, filter, update)
	return err
}
