package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id" swaggertype:"string"`
	Title       string             `bson:"title" json:"title"`
	Duration    int                `bson:"duration" json:"duration"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at" swaggertype:"string"`
	UpdatedAt   primitive.DateTime `bson:"updated_at" json:"updated_at" swaggertype:"string"`
}
