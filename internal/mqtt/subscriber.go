package mqtt

import (
	"log"

	"fleet-management/internal/config"
	"fleet-management/internal/services"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const topicPattern = "/fleet/vehicle/+/location"

type Subscriber struct {
	client  mqtt.Client
	handler *services.LocationService
}

func NewSubscriber(cfg *config.Config, handler *services.LocationService) (*Subscriber, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("fleet-backend-subscriber").
		SetAutoReconnect(true)
	if cfg.MQTTUser != "" {
		opts.SetUsername(cfg.MQTTUser)
		opts.SetPassword(cfg.MQTTPassword)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	sub := &Subscriber{client: client, handler: handler}

	token := client.Subscribe(topicPattern, 1, sub.messageHandler)
	if token.Wait() && token.Error() != nil {
		client.Disconnect(250)
		return nil, token.Error()
	}

	return sub, nil
}

func (s *Subscriber) messageHandler(client mqtt.Client, msg mqtt.Message) {
	if err := s.handler.ProcessMQTTMessage(msg.Payload()); err != nil {
		log.Printf("mqtt message processing error: %v", err)
	}
}

func (s *Subscriber) Close() {
	s.client.Disconnect(250)
}
