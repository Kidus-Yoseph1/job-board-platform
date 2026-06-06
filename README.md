# Job Board Platform (Microservices Architecture)

A modern, high-performance, and resilient microservices-based job board platform built in Go. This repository features two main services that communicate asynchronously via Redis Pub/Sub for real-time notifications and synchronously via gRPC for user metadata resolution.

---

## Services Overview

### 1. [Job Service](./job-service)
Handles core business logic including user registration, authentication, job postings, and applications.
*   **API Protocol:** HTTP/REST (Port 8080)
*   **gRPC Protocol:** Server listening on Port 50051
*   **Storage:** PostgreSQL (persisting users, jobs, companies, and applications)
*   **Caching:** Redis (caching job listings and details)
*   **Messaging:** Publishes events (e.g., `job_applied`, `application_status_changed`) to Redis Pub/Sub.

### 2. [Notification Service](./notification-service)
Manages real-time alerts, email queues, and historical notifications.
*   **API Protocol:** HTTP/REST (Port 8081)
*   **WebSocket Protocol:** Live alerts on `/ws`
*   **gRPC Protocol:** Client communicating with Job Service on Port 50051
*   **Storage:** PostgreSQL (persisting notification history and email jobs)
*   **Messaging:** Subscribes to Redis Pub/Sub events.
*   **Workers:** Background worker that periodically polls and sends queued emails via SMTP.

---

## Local Setup & Running

### Prerequisites
*   Go 1.22+
*   Docker & Docker Compose
*   `sqlc` (for SQL database query generation)
*   `protoc` (for gRPC code generation)

### 1. Start Infrastructure (PostgreSQL & Redis)
Ensure Docker is running and execute:
```bash
docker compose up -d
```
This runs:
*   **PostgreSQL** on `localhost:5434` (DBs: `jobboard` & `notifications`)
*   **Redis** on `localhost:6379`

### 2. Run Database Migrations
Make sure migrations are run for both databases:
*   For the job database: Apply the files inside `job-service/migrations`
*   For the notification database: Apply the files inside `notification-service/migrations`

### 3. Start Job Service
```bash
cd job-service
# Configure .env (ensure DB_PORT=5434 and GRPC_PORT=50051)
go run cmd/server/main.go
```

### 4. Start Notification Service
```bash
cd ../notification-service
# Configure .env (ensure DB_PORT=5434 and JOB_SERVICE_GRPC_ADDR=localhost:50051)
go run cmd/server/main.go
```

---

## Go Workspace (`go.work`)
This project utilizes a Go multi-module workspace (`go.work`) to cleanly manage dependencies and local import paths across `job-service`, `notification-service`, and the shared `proto` module. 

To work with this structure:
*   Make sure `go.work` is present in the root.
*   Run `go build` or `go test` from within individual module directories, and the local dependencies will resolve automatically.

---

## Shared Protocols (`proto`)
If you modify the gRPC service contracts under `./proto/jobboard.proto`, regenerate the Go proto/gRPC files by running the following command from the root directory:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/jobboard.proto
```