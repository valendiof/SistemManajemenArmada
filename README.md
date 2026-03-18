# Sistem Manajemen Armada (Fleet Management System)

Sistem backend untuk memantau lokasi kendaraan secara real-time menggunakan MQTT, REST API, PostgreSQL, RabbitMQ, dan geofencing.

## Prasyarat

- Docker dan Docker Compose
- Go 1.21+ (opsional, untuk menjalankan tanpa Docker)

## Quick Start dengan Docker

```bash
docker compose up -d
```

Semua service akan berjalan. Tunggu beberapa detik hingga backend siap menerima koneksi.

### Verifikasi

- **REST API:** http://localhost:8080
- **RabbitMQ Management:** http://localhost:15672 (fleet/rabbitpass)
- **PostgreSQL:** localhost:5432 (fleet/fleetpass, database: fleet)
- **MQTT:** localhost:1883

## Autentikasi

REST API memakai API key. Default dari docker-compose:

- `API_KEY`: Dikonfigurasi melalui file `config.env` (Default: `K8P2N5W9Q4X1V7M6`)

Anda bisa kirim kredensial via salah satu header:

- `Authorization: Bearer <API_KEY>`
- `X-API-Key: <API_KEY>`

## Service dan Port

| Service | Port | Deskripsi |
|---------|------|-----------|
| app-backend | 8080 | REST API + MQTT subscriber |
| app-worker | - | Consumer RabbitMQ (geofence alerts) |
| app-mock | - | Mock publisher MQTT (setiap 2 detik) |
| postgres | 5432 | Database |
| rabbitmq | 5672, 15672 | Message broker |
| eclipse-mosquitto | 1883 | MQTT broker |

## API Endpoints

### GET /vehicles/:vehicle_id/location

Mengembalikan lokasi terakhir kendaraan.

**Contoh:**
```bash
curl -H "Authorization: Bearer K8P2N5W9Q4X1V7M6" http://localhost:8080/vehicles/vehicle-001/location
```

### GET /vehicles/:vehicle_id/history?start=X&end=Y

Mengembalikan array lokasi dalam rentang timestamp (Unix epoch).

**Contoh:**
```bash
curl -H "Authorization: Bearer K8P2N5W9Q4X1V7M6" "http://localhost:8080/vehicles/vehicle-001/history?start=1710000000&end=1710100000"
```

### GET /vehicles/:vehicle_id/history/today

Mengembalikan array lokasi dalam rentang hari ini.

**Contoh:**
```bash
curl -H "Authorization: Bearer K8P2N5W9Q4X1V7M6" "http://localhost:8080/vehicles/vehicle-001/history/today"
```

## Alur Testing

1. Jalankan `docker compose up -d`
2. Mock publisher mengirim data ke MQTT setiap 2 detik (vehicle_id: vehicle-001)
3. Backend menerima, menyimpan ke PostgreSQL, dan jika dalam radius 50m dari titik pusat (-6.2000, 106.8166) mempublish ke RabbitMQ
4. Worker mencetak alert geofence ke console: `docker compose logs -f app-worker`
5. Cek lokasi terakhir: `curl -H "Authorization: Bearer K8P2N5W9Q4X1V7M6" http://localhost:8080/vehicles/vehicle-001/location`
6. Cek history: `curl -H "Authorization: Bearer K8P2N5W9Q4X1V7M6" "http://localhost:8080/vehicles/vehicle-001/history?start=0&end=9999999999"`

## Setelah `docker compose up -d`, apa yang dilakukan?

1. Pastikan semua container healthy/running:

```bash
docker compose ps
```

2. Lihat log service utama:

```bash
docker compose logs -f app-backend
```

3. Lihat log worker untuk geofence event:

```bash
docker compose logs -f app-worker
```

4. Gunakan Postman collection di folder `postman/` atau gunakan curl dengan header API key.

## Mengecek Database PostgreSQL

Menjalankan query dari dalam container postgres:

```bash
docker compose exec postgres psql -U fleet -d fleet -c "SELECT vehicle_id, latitude, longitude, timestamp FROM vehicle_locations ORDER BY timestamp DESC LIMIT 10;"
```

Melihat jumlah data:

```bash
docker compose exec postgres psql -U fleet -d fleet -c "SELECT count(*) FROM vehicle_locations;"
```

Jika ingin akses dari GUI (DBeaver/pgAdmin), konek ke:

- Host: `localhost`
- Port: `5432`
- User: `fleet`
- Password: `fleetpass`
- Database: `fleet`

## Menjalankan Tanpa Docker

Pastikan PostgreSQL, RabbitMQ, dan Mosquitto sudah berjalan lokal.

```bash
export DB_URL="postgres://fleet:fleetpass@localhost:5432/fleet?sslmode=disable"
export MQTT_BROKER="tcp://localhost:1883"
export MQTT_USER="fleet"
export MQTT_PASSWORD="mqttpass"
export RABBITMQ_URL="amqp://fleet:rabbitpass@localhost:5672/"
# Clone repository dan buat file config.env:
# API_KEY=K8P2N5W9Q4X1V7M6

go run ./cmd/backend
```

Di terminal lain:
```bash
go run ./cmd/worker
```

Di terminal lain:
```bash
go run ./cmd/mock
```

## Struktur Project

```
├── cmd/
│   ├── backend/     Server utama (REST API + MQTT)
│   ├── worker/      RabbitMQ consumer
│   └── mock/        Mock MQTT publisher
├── internal/
│   ├── config/      Konfigurasi
│   ├── handlers/    Gin handlers
│   ├── models/      Data models
│   ├── repository/  PostgreSQL repository
│   ├── services/    Business logic (Haversine, RabbitMQ)
│   └── mqtt/        MQTT subscriber
├── migrations/      SQL schema
├── Dockerfile
└── docker-compose.yml
```

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| DB_URL | postgres://fleet:fleetpass@localhost:5432/fleet?sslmode=disable | Connection string PostgreSQL |
| MQTT_BROKER | tcp://localhost:1883 | URL broker MQTT |
| MQTT_USER | (empty) | Username MQTT (kalau broker butuh auth) |
| MQTT_PASSWORD | (empty) | Password MQTT |
| RABBITMQ_URL | amqp://fleet:rabbitpass@localhost:5672/ | URL RabbitMQ |
| API_KEY | (empty) | API key untuk REST API |
| PORT | 8080 | Port REST API |
| MOCK_VEHICLE_ID | vehicle-001 | Vehicle ID untuk mock publisher |

## Postman

Import file berikut di Postman:

- `postman/FleetManagement.postman_collection.json`
- `postman/FleetManagement.postman_environment.json`
