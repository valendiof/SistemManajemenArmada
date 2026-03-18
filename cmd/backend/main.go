package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"fleet-management/internal/config"
	"fleet-management/internal/handlers"
	"fleet-management/internal/mqtt"
	"fleet-management/internal/repository"
	"fleet-management/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	migratePath := os.Getenv("MIGRATIONS_PATH")
	if migratePath == "" {
		migratePath = "migrations/001_create_vehicle_locations.sql"
	}
	migration, err := os.ReadFile(migratePath)
	if err == nil {
		if _, err := db.Exec(string(migration)); err != nil {
			log.Printf("migration warning: %v", err)
		}
	}

	rabbitPub, err := services.NewRabbitMQPublisher(cfg)
	if err != nil {
		log.Fatalf("failed to create rabbitmq publisher: %v", err)
	}
	defer rabbitPub.Close()

	locRepo := repository.NewVehicleLocationRepository(db)
	locService := services.NewLocationService(locRepo, rabbitPub)

	subscriber, err := mqtt.NewSubscriber(cfg, locService)
	if err != nil {
		log.Fatalf("failed to create mqtt subscriber: %v", err)
	}
	defer subscriber.Close()

	vehicleHandler := handlers.NewVehicleHandler(locService)
	authMiddleware := handlers.NewAuthMiddleware(cfg.APIKey)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		page := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Sistem Manajemen Armada</title>
  <style>
    html, body {
      height: 100%;
      margin: 0;
    }
    body {
      display: flex;
      align-items: center;
      justify-content: center;
      background: #ffffff;
      color: #111;
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      text-rendering: optimizeLegibility;
    }
    .container {
      text-align: center;
      padding: 1rem;
    }
    h1 {
      font-size: 3rem;
      margin: 0;
    }
    .version {
      margin-top: 0.5rem;
      font-size: 1.25rem;
      color: #444;
    }
    .env {
      margin-top: 1rem;
      font-size: 0.9rem;
      color: #666;
      letter-spacing: 0.15em;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Sistem Manajemen Armada</h1>
    <div class="version">Version 1.0.0</div>
    <div class="env">~~~~~~</div>
  </div>
</body>
</html>`

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
	})
	vehicles := r.Group("/vehicles", authMiddleware.RequireAPIKey())
	vehicles.GET("/:vehicle_id/location", vehicleHandler.GetLocation)
	vehicles.GET("/:vehicle_id/history", vehicleHandler.GetHistory)
	vehicles.GET("/:vehicle_id/history/today", vehicleHandler.GetHistoryToday)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("backend starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
