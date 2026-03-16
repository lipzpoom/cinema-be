package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gin-quickstart/internal/queue"

	amqp "github.com/rabbitmq/amqp091-go"
)

type BookingSuccessPayload struct {
	BookingID string `json:"booking_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func PublishBookingSuccess(payload BookingSuccessPayload) error {
	if queue.RabbitChannel == nil {
		log.Println("⚠️ RabbitMQ channel is nil, skipping message publish")
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = queue.RabbitChannel.PublishWithContext(ctx,
		"booking_events", // exchange
		"",               // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Printf("❌ Failed to publish message: %v", err)
		return err
	}

	log.Printf("📥 Published message to queue: %s", body)
	return nil
}
