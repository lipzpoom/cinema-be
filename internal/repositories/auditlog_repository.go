package repositories

import (
	"context"
	"gin-quickstart/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepository struct {
	Collection *mongo.Collection
}

func NewAuditLogRepository(db *mongo.Database) *AuditLogRepository {
	return &AuditLogRepository{
		Collection: db.Collection("audit_logs"),
	}
}

func (r *AuditLogRepository) GetAll(ctx context.Context) ([]models.AuditLog, error) {
	var items []models.AuditLog
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	// ✅ Pass bson.M{} as the filter (to find all) and findOptions as the options
	cursor, err := r.Collection.Find(ctx, bson.M{}, findOptions)

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	if items == nil {
		items = []models.AuditLog{}
	}

	return items, nil
}
