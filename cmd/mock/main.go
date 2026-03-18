package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"fleet-management/internal/config"
	"fleet-management/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	cfg := config.Load()

	vehicleID := os.Getenv("MOCK_VEHICLE_ID")
	if vehicleID == "" {
		vehicleID = "vehicle-001"
	}

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID("fleet-mock-publisher").
		SetAutoReconnect(true)
	if cfg.MQTTUser != "" {
		opts.SetUsername(cfg.MQTTUser)
		opts.SetPassword(cfg.MQTTPassword)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect to mqtt: %v", token.Error())
	}
	defer client.Disconnect(250)

	topic := "/fleet/vehicle/" + vehicleID + "/location"
	centerLat := -6.2000
	centerLng := 106.8166

	log.Printf("publishing mock locations to %s every 2s (vehicle_id=%s)", topic, vehicleID)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		offsetLat := (rand.Float64() - 0.5) * 0.001
		offsetLng := (rand.Float64() - 0.5) * 0.001
		loc := models.VehicleLocation{
			VehicleID: vehicleID,
			Latitude:  centerLat + offsetLat,
			Longitude: centerLng + offsetLng,
			Timestamp: time.Now().Unix(),
		}

		payload, _ := json.Marshal(loc)
		token := client.Publish(topic, 1, false, payload)
		if token.Wait() && token.Error() != nil {
			log.Printf("publish error: %v", token.Error())
			continue
		}
		log.Printf("published: %s", string(payload))
	}
}
