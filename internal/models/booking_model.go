package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingStatus string

const (
	BookingPending BookingStatus = "PENDING"
	BookingSuccess BookingStatus = "SUCCESS"
	BookingFailed  BookingStatus = "FAILED"
)

type Booking struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	Seats      []string           `bson:"seats" json:"seats"`
	Status     BookingStatus      `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
