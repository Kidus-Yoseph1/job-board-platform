# Notification Service

A dedicated microservice for managing real-time notifications, notification history, and asynchronous email delivery. It subscribes to Redis Pub/Sub events emitted by the `job-service`, persists notifications to PostgreSQL, pushes live alerts over WebSocket, and dispatches emails via a background worker.

---

## Ports

| Protocol  | Port  | Purpose                                       |
|-----------|-------|-----------------------------------------------|
| HTTP      | 8081  | REST API for notification history             |
| WebSocket | 8081  | Real-time alerts served on `GET /ws`          |

---

## Environment Variables (`.env`)

| Variable                | Default               | Description                                                    |
|-------------------------|-----------------------|----------------------------------------------------------------|
| `PORT`                  | `8081`                | HTTP/WebSocket server port                                     |
| `JOB_SERVICE_GRPC_ADDR` | `localhost:50051`     | gRPC address of the job-service for user metadata lookup       |
| `DB_HOST`               | `localhost`           | PostgreSQL host                                                |
| `DB_PORT`               | `5434`                | PostgreSQL port                                                |
| `DB_USER`               | `postgres`            | PostgreSQL username                                            |
| `DB_PASSWORD`           | `postgres`            | PostgreSQL password                                            |
| `DB_NAME`               | `notifications`       | PostgreSQL database name                                       |
| `DB_SSLMODE`            | `disable`             | PostgreSQL SSL mode                                            |
| `REDIS_ADDR`            | `localhost:6379`      | Redis server address                                           |
| `JWT_SECRET`            | `secret`              | Must match the JWT secret used in the job-service              |
| `SMTP_HOST`             | `localhost`           | SMTP server host. Set to `localhost` with no username to simulate in dev |
| `SMTP_PORT`             | `587`                 | SMTP server port                                               |
| `SMTP_USERNAME`         | *(empty)*             | SMTP username (leave empty for dev simulation mode)            |
| `SMTP_PASSWORD`         | *(empty)*             | SMTP password                                                  |
| `SMTP_FROM_EMAIL`       | `no-reply@jobboard.com` | Sender address shown in outgoing emails                      |

> **Development Email Mode:** If `SMTP_HOST` is `localhost` and `SMTP_USERNAME` is empty, the email worker will simulate a successful send without connecting to any SMTP server. This is safe and intentional for local development.

---

## API Endpoints

All endpoints require a valid JWT issued by the `job-service`. Pass it in the `Authorization: Bearer <token>` header.

### Notifications

| Method | Path                       | Description                                          |
|--------|----------------------------|------------------------------------------------------|
| GET    | `/notifications`           | Fetch all notifications for the authenticated user   |
| PUT    | `/notifications/read`      | Mark all notifications as read for the current user  |
| PUT    | `/notifications/read/:id`  | Mark a single notification as read by its ID         |

**`GET /notifications` Response:**
```json
{
  "status": "success",
  "data": {
    "notifications": [
      {
        "id": "b92e41b2-81be-4620-855d-d05fd9ba5bb6",
        "user_id": "f8d8b9fb-5429-4283-bb88-501d1fdfee5a",
        "type": "job_applied",
        "title": "New Application Received",
        "body": "A new candidate has applied for your job posting: Backend Engineer",
        "read": false,
        "created_at": "2026-06-05T13:55:52Z"
      }
    ]
  }
}
```

---

### WebSocket — Real-Time Notifications

**Endpoint:** `GET /ws`
**Authentication:** JWT must be passed as a query parameter:
```
ws://localhost:8081/ws?token=<your_jwt>
```

Once connected, the service pushes live notification payloads to the client's browser whenever a relevant event occurs. No messages need to be sent from the client.

**Incoming Push Message Format:**
```json
{
  "type": "job_applied",
  "message": "A new candidate has applied for your job posting: Backend Engineer",
  "data": {
    "id": "b92e41b2-81be-4620-855d-d05fd9ba5bb6",
    "user_id": "...",
    "type": "job_applied",
    "title": "New Application Received",
    "body": "...",
    "read": false,
    "created_at": "2026-06-05T13:55:52Z"
  }
}
```

---

## Internal Event Pipeline

This section explains the internal flow after a Redis event is received.

### Step 1: Redis Subscriber (`internal/subscriber/redis_subscriber.go`)
*   Permanently subscribes to the `job_events` channel in Redis.
*   Runs in a background goroutine from startup.
*   Deserializes incoming JSON messages into a typed `RedisEvent` struct.
*   Validates the `target_user_id` is a valid UUID before proceeding.

### Step 2: Persist Notification History (`db/generated/notifications.sql.go`)
*   Saves a permanent record to the `notifications` PostgreSQL table.
*   Fields: `user_id`, `type`, `title`, `body`, `read` (default: `false`).
*   This record is what users see when they call `GET /notifications`.

### Step 3: WebSocket Push (`internal/websocket/hub.go`)
*   Calls `hub.SendNotification(targetUUID, payload)`.
*   The Hub checks its in-memory map of connected WebSocket clients for the `targetUUID`.
*   If the user is currently online, it writes the JSON payload into that client's `send` channel.
*   The `Client.writePump()` goroutine reads from the channel and transmits the data over the open TCP socket to the browser immediately.

### Step 4: gRPC Metadata Enrichment (`internal/grpc/client.go`)
*   Calls `GetUserDetails(target_user_id)` on the `job-service`'s gRPC server (port 50051).
*   Retrieves the recipient's email address without directly querying the `job-service`'s database.

### Step 5: Email Job Queue (`db/generated/email_jobs.sql.go`)
*   Inserts a new record into the `email_jobs` PostgreSQL table with status `pending`.
*   Fields: `to_email`, `subject`, `body`.

### Step 6: Background Email Worker (`internal/worker/email_worker.go`)
*   Polls the `email_jobs` table every **5 seconds** for `pending` or `failed` records with fewer than **3 attempts**.
*   Sends each email via the configured SMTP server.
*   Updates the record status to `sent` on success, or `failed` on error (incrementing the `attempts` counter).
*   Maximum retry attempts: **3**.

---

## Project Structure

```
notification-service/
├── cmd/server/main.go              # Entrypoint: boots HTTP, WebSocket Hub, Redis subscriber, email worker
├── db/
│   ├── migrations/                 # SQL migration files (schema creation)
│   ├── queries/                    # Raw SQL queries consumed by sqlc
│   └── generated/                  # Auto-generated Go code from sqlc (do not edit)
├── internal/
│   ├── config/config.go            # Environment variable loading via Viper (includes SMTP config)
│   ├── grpc/client.go              # gRPC client wrapper: calls job-service's GetUserDetails RPC
│   ├── handler/                    # HTTP handlers (notification history, WebSocket upgrade)
│   ├── middleware/                 # JWT auth middleware, CORS
│   ├── repository/                 # Database access layer (notifications, email jobs)
│   ├── service/notification_service.go # Business logic for listing and marking notifications
│   ├── subscriber/redis_subscriber.go  # Redis Pub/Sub listener and event handler
│   ├── websocket/
│   │   ├── hub.go                  # Central WebSocket Hub: manages all active connections
│   │   └── client.go               # Per-connection WebSocket client: read/write pumps
│   └── worker/email_worker.go      # Background goroutine: polls & sends queued emails via SMTP
├── pkg/
│   ├── logger/logger.go            # Zap-based structured logger with file and console sinks
│   └── response/                   # Standardized JSON success/error response helpers
└── go.mod
```

---

## Logs

All logs are written to the `logs/` directory at runtime:

| File             | Contents                                    |
|------------------|---------------------------------------------|
| `logs/app.log`   | All log levels (Debug, Info, Warn, Error)   |
| `logs/error.log` | Error level only for quick triage           |

Both log files use plain-text format (no color codes) and are safe to `grep` or pipe into log aggregators.
