package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"fleet-management/internal/config"
	"fleet-management/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName = "fleet.events"
	queueName    = "geofence_alerts"
)

type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.Config
}

func NewRabbitMQPublisher(cfg *config.Config) (*RabbitMQPublisher, error) {
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
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(q.Name, "geofence.entry", exchangeName, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQPublisher{conn: conn, channel: ch, config: cfg}, nil
}

func (p *RabbitMQPublisher) PublishGeofenceEvent(event *models.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		exchangeName,
		"geofence.entry",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *RabbitMQPublisher) PublishIfInGeofence(loc *models.VehicleLocation) {
	dist := HaversineDistanceMeters(
		p.config.GeofenceLat, p.config.GeofenceLng,
		loc.Latitude, loc.Longitude,
	)
	if dist > p.config.GeofenceMeters {
		return
	}

	event := &models.GeofenceEvent{
		VehicleID: loc.VehicleID,
		Event:     "geofence_entry",
		Location: models.LatLng{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		},
		Timestamp: loc.Timestamp,
		DateTime:  time.Unix(loc.Timestamp, 0).In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05"),
	}

	if err := p.PublishGeofenceEvent(event); err != nil {
		log.Printf("failed to publish geofence event: %v", err)
	}
}
