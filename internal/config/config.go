package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL          string
	MQTTBroker     string
	MQTTUser       string
	MQTTPassword   string
	RabbitMQURL    string
	APIKey         string
	GeofenceLat    float64
	GeofenceLng    float64
	GeofenceMeters float64
}

func Load() *Config {
	_ = godotenv.Load("config.env")

	cfg := &Config{
		DBURL:          getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/fleet?sslmode=disable"),
		MQTTBroker:     getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTUser:       getEnv("MQTT_USER", ""),
		MQTTPassword:   getEnv("MQTT_PASSWORD", ""),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		APIKey:         getEnv("API_KEY", ""),
		GeofenceLat:    -6.2000,
		GeofenceLng:    106.8166,
		GeofenceMeters: 50,
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
