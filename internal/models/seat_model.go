package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	SeatNumber string             `bson:"seat_number" json:"seat_number"`
	Status     SeatStatus         `bson:"status" json:"status"`
}
