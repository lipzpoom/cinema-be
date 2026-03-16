package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gin-quickstart/internal/events"
	"gin-quickstart/internal/models"
	"gin-quickstart/internal/queue"
	"gin-quickstart/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartLogWorker(db *mongo.Database, wsManager *websocket.AuditLogManager) {
msgs, err := queue.RabbitChannel.Consume(
"audit_log_queue", // queue name
"log_worker",      // consumer name
true,              // auto-ack
false,             // exclusive
false,             // no-local
false,             // no-wait
nil,               // args
)
if err != nil {
log.Fatalf("? Failed to register log consumer: %v", err)
}

auditCollection := db.Collection("audit_logs")

go func() {
for d := range msgs {
var payload events.AuditLogPayload
if err := json.Unmarshal(d.Body, &payload); err != nil {
log.Printf("? Error decoding log message: %v", err)
continue
}

log.Printf("📥 [Log Worker] Received AuditLog from RabbitMQ -> Event: %s, User: %s. Saving to MongoDB...", payload.Event, payload.UserID)

userID, _ := primitive.ObjectIDFromHex(payload.UserID)

auditLog := models.AuditLog{
ID:         primitive.NewObjectID(),
Event:      payload.Event,
UserID:     userID,
Value:      payload.Value, 
CreatedAt:  payload.Timestamp,
}

if auditLog.CreatedAt.IsZero() {
auditLog.CreatedAt = time.Now()
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
_, err := auditCollection.InsertOne(ctx, auditLog)
cancel()

if err != nil {
log.Printf("? [Log Worker] Failed to save audit log: %v", err)
} else {
log.Printf("?? [Log Worker] Audit Log saved for Event: %s, User: %s", payload.Event, payload.UserID)
if wsManager != nil {
wsManager.Broadcast <- auditLog
}
}
}
}()

log.Println("?? Log Worker started, waiting for messages...")
}
