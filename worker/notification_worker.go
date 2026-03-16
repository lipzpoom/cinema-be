package worker

import (
	"encoding/json"
	"log"

	"gin-quickstart/internal/events"
	"gin-quickstart/internal/queue"
)

func StartNotificationWorker() {
	msgs, err := queue.RabbitChannel.Consume(
		"notification_queue",  // queue name
		"notification_worker", // consumer name
		true,                  // auto-ack (บอก RabbitMQ ว่ารับข้อความแล้วอัตโนมัติ)
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		log.Fatalf("❌ Failed to register notification consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var payload events.BookingSuccessPayload
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("❌ Error decoding notification message: %v", err)
				continue
			}

			// จำลองการใช้เวลาส่ง Email หรือ SMS
			log.Printf("📧 [Notification Worker] Sending Email/SMS for Booking ID: %s to User: %s (Status: %s)", payload.BookingID, payload.UserID, payload.Status)
		}
	}()

	log.Println("👷 Notification Worker started, waiting for messages...")
}
