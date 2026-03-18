package main

import (
	"encoding/json"
	"log"
	"time"

	"fleet-management/internal/config"
	"fleet-management/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "geofence_alerts"

func main() {
	cfg := config.Load()

	var conn *amqp.Connection
	var err error
	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	_ = ch.ExchangeDeclare("fleet.events", "topic", true, false, false, false, nil)
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}
	_ = ch.QueueBind(q.Name, "geofence.entry", "fleet.events", false, nil)

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to register consumer: %v", err)
	}

	log.Printf("worker listening on queue %s", queueName)

	for msg := range msgs {
		var event models.GeofenceEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("failed to decode message: %v", err)
			continue
		}
		log.Printf("geofence alert: vehicle_id=%s event=%s lat=%.6f lng=%.6f timestamp=%d",
			event.VehicleID, event.Event, event.Location.Latitude, event.Location.Longitude, event.Timestamp)
	}
}
