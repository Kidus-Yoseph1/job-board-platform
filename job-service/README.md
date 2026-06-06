# Job Service

The core backend microservice of the Job Board Platform. It handles user authentication, job posting management, company profiles, and job applications. When key application events occur, it publishes events to Redis Pub/Sub and exposes a gRPC server to serve metadata to other microservices.

---

## Ports

| Protocol | Port  | Purpose                                |
|----------|-------|----------------------------------------|
| HTTP     | 8080  | REST API for all clients               |
| gRPC     | 50051 | Internal metadata server for services |

---

## Environment Variables (`.env`)

| Variable      | Default       | Description                                     |
|---------------|---------------|-------------------------------------------------|
| `PORT`        | `8080`        | HTTP server port                                |
| `GRPC_PORT`   | `50051`       | gRPC server port                                |
| `DB_HOST`     | `localhost`   | PostgreSQL host                                 |
| `DB_PORT`     | `5434`        | PostgreSQL port                                 |
| `DB_USER`     | `postgres`    | PostgreSQL username                             |
| `DB_PASSWORD` | `postgres`    | PostgreSQL password                             |
| `DB_NAME`     | `jobboard`    | PostgreSQL database name                        |
| `DB_SSLMODE`  | `disable`     | PostgreSQL SSL mode                             |
| `REDIS_ADDR`  | `localhost:6379` | Redis server address                         |
| `JWT_SECRET`  | `secret`      | Secret key for signing and verifying JWT tokens |

---

## API Endpoints

### Auth

| Method | Path             | Auth Required | Role | Description              |
|--------|------------------|---------------|------|--------------------------|
| POST   | `/auth/register` | ❌            | —    | Register a new user      |
| POST   | `/auth/login`    | ❌            | —    | Login and receive a JWT  |

**Register Request Body:**
```json
{
  "full_name":"user name",
  "email": "user@example.com",
  "password": "securepassword",
  "role": "candidate"
}
```
> `role` must be either `"candidate"` or `"company"`.

**Login Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

---

### Jobs (Public)

| Method | Path                   | Auth Required | Role | Description                  |
|--------|------------------------|---------------|------|------------------------------|
| GET    | `/jobs`                | ❌            | —    | List all active jobs         |
| GET    | `/job/:id`             | ❌            | —    | Get a specific job by its ID |
| GET    | `/jobs/company/:id`    | ❌            | —    | List jobs by a company ID    |

**Query Parameters for `GET /jobs`:**
| Param    | Type   | Default | Description              |
|----------|--------|---------|--------------------------|
| `limit`  | int    | 20      | Number of results        |
| `offset` | int    | 0       | Pagination offset        |

---

### Jobs (Company Only)

> Requires a valid JWT with `role: company`.

| Method | Path                 | Auth Required | Role    | Description             |
|--------|----------------------|---------------|---------|-------------------------|
| POST   | `/create_job`        | ✅            | company | Create a new job posting |
| PUT    | `/update_job/:id`    | ✅            | company | Update a job's status    |
| DELETE | `/delete_job/:id`    | ✅            | company | Soft-delete a job        |

**Create Job Request Body:**
```json
{
  "title": "Backend Engineer",
  "description": "Build APIs in Go...",
  "location": "Remote",
  "salary_min": 50000,
  "salary_max": 90000,
  "job_type": "full_time"
}
```

---

### Applications (Candidate Only)

> Requires a valid JWT with `role: candidate`.

| Method | Path                | Auth Required | Role      | Description                         |
|--------|---------------------|---------------|-----------|-------------------------------------|
| POST   | `/apply/:id`        | ✅            | candidate | Apply to a job (`:id` = job ID)     |
| GET    | `/application/:id`  | ✅            | candidate | Get a specific application by ID    |
| GET    | `/application`      | ✅            | candidate | List all of the authenticated user's applications |

**Apply Request Body:**
```json
{
  "cover_letter": "I am very interested in this role..."
}
```

---

### Applications (Company Only)

> Requires a valid JWT with `role: company`.

| Method | Path                          | Auth Required | Role    | Description                             |
|--------|-------------------------------|---------------|---------|-----------------------------------------|
| GET    | `/application/job/:job_id`    | ✅            | company | List all applications for a job posting |
| PUT    | `/application/update/:id`     | ✅            | company | Update an application's status          |

**Update Application Status Body:**
```json
{
  "status": "accepted"
}
```
> `status` can be `"pending"`, `"reviewed"`, `"accepted"`, or `"rejected"`.

---

## Redis Pub/Sub Events Published

When a candidate applies or a company updates an application status, the service publishes a structured event to the `job_events` Redis channel. The `notification-service` subscribes to this channel.

### `job_applied` Event
Published when a candidate successfully submits an application.
```json
{
  "event_type": "job_applied",
  "payload": {
    "target_user_id": "<company_owner_user_uuid>",
    "title": "New Application Received",
    "message": "A new candidate has applied for your job posting: Backend Engineer"
  }
}
```

### `application_status_changed` Event
Published when a company updates the status of an application.
```json
{
  "event_type": "application_status_changed",
  "payload": {
    "target_user_id": "<candidate_user_uuid>",
    "title": "Application Status Updated",
    "message": "Your application for 'Backend Engineer' has been updated to: accepted"
  }
}
```

---

## gRPC Server (Port 50051)

Exposes metadata lookup endpoints consumed by the `notification-service`. The service contract is defined in `../proto/jobboard.proto`.

| RPC Method        | Request            | Response            | Description                        |
|-------------------|--------------------|---------------------|------------------------------------|
| `GetJobDetails`   | `{ job_id }`       | Job title, description, company name | Fetch job details by ID |
| `GetUserDetails`  | `{ user_id }`      | Full name, email    | Fetch user profile by ID           |

---

## Project Structure

```
job-service/
├── cmd/server/main.go          # Entrypoint: boots HTTP, gRPC, Redis, and DB
├── db/
│   ├── migrations/             # SQL migration files (schema creation)
│   ├── queries/                # Raw SQL queries consumed by sqlc
│   └── generated/              # Auto-generated Go code from sqlc (do not edit)
├── internal/
│   ├── cache/redis.go          # Redis client wrapper with Get, Set, Del, and Publish
│   ├── config/config.go        # Environment variable loading via Viper
│   ├── domain/errors.go        # Typed application errors (NotFound, Forbidden, Internal)
│   ├── grpc/server.go          # gRPC server implementing GetJobDetails & GetUserDetails
│   ├── handler/                # HTTP request handlers (auth, jobs, applications)
│   ├── middleware/             # JWT auth middleware, role middleware, CORS
│   ├── repository/             # Database access layer (user, job, company, application)
│   ├── routes/routes.go        # Route registration for all HTTP endpoints
│   └── service/                # Business logic layer (auth, job, application)
├── pkg/
│   ├── logger/logger.go        # Zap-based structured logger with file and console sinks
│   └── response/               # Standardized JSON success/error response helpers
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
