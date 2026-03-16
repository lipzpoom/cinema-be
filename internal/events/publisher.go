package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gin-quickstart/internal/queue"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AuditLogPayload struct {
	Event     string    `json:"event"`
	UserID    string    `json:"user_id"`
	Value     string    `json:"value,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func PublishAuditLog(payload AuditLogPayload) error {
if queue.RabbitChannel == nil {
log.Println("?? RabbitMQ channel is nil, skipping message publish")
return nil
}

body, err := json.Marshal(payload)
if err != nil {
return err
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err = queue.RabbitChannel.PublishWithContext(ctx,
"audit_events",   // exchange
"",               // routing key
false,            // mandatory
false,            // immediate
amqp.Publishing{
ContentType: "application/json",
Body:        body,
})

if err != nil {
log.Printf("? Failed to publish audit log message: %v", err)
return err
}

log.Printf("?? [Publisher] Sent AuditLog to RabbitMQ -> Event: %s, User: %s", payload.Event, payload.UserID)
return nil
}
