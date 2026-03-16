package queue

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitConn *amqp.Connection
var RabbitChannel *amqp.Channel

func InitRabbitMQ(dsn string) error {
	var err error
	RabbitConn, err = amqp.Dial(dsn)
	if err != nil {
		return err
	}

	RabbitChannel, err = RabbitConn.Channel()
	if err != nil {
		return err
	}

	// Exchange สำหรับกระจาย Event (Fanout)
	err = RabbitChannel.ExchangeDeclare(
		"booking_events", // name
		"fanout",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}

	err = RabbitChannel.ExchangeDeclare(
		"audit_events", // name
		"fanout",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	// Queue 1: สำหรับแจ้งเตือน (Notification)
	_, err = RabbitChannel.QueueDeclare("notification_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Queue 2: สำหรับเก็บ Log (Audit)
	_, err = RabbitChannel.QueueDeclare("audit_log_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// ผูก (Bind) Queue เข้ากับ Exchange
	err = RabbitChannel.QueueBind("notification_queue", "", "booking_events", false, nil)
	if err != nil {
		return err
	}
	err = RabbitChannel.QueueBind("audit_log_queue", "", "audit_events", false, nil)
	if err != nil {
		return err
	}

	log.Println("✅ RabbitMQ Connected, Exchange and Queues Declared")
	return nil
}

func CloseRabbitMQ() {
	if RabbitChannel != nil {
		RabbitChannel.Close()
	}
	if RabbitConn != nil {
		RabbitConn.Close()
	}
}
