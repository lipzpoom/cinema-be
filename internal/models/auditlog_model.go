package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Event      string             `bson:"event" json:"event"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Value      string             `bson:"value" json:"value"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
