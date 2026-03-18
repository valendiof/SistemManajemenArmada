FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o backend ./cmd/backend
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker
RUN CGO_ENABLED=0 go build -o mock ./cmd/mock

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/backend .
COPY --from=builder /build/worker .
COPY --from=builder /build/mock .
COPY --from=builder /build/migrations ./migrations
