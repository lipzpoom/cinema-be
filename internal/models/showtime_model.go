package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Showtime struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieID       primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	StartTime     time.Time          `bson:"start_time" json:"start_time"`
	EndTime       time.Time          `bson:"end_time" json:"end_time"`
	Price         float64            `bson:"price" json:"price"`
	TheaterNumber string             `bson:"theater_number" json:"theater_number"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}
