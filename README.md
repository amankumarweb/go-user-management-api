# User API — Go REST Service

A production-quality RESTful API built in **Go** for managing users with `name` and `dob` (date of birth). The API dynamically calculates and returns each user's **age** — it is never stored in the database.

---

## Architecture Overview

```
Client ──▶ Fiber (HTTP) ──▶ Handler ──▶ Service ──▶ Repository ──▶ SQLC / PostgreSQL
                │
           Middleware
         (Request ID,
          Logger, CORS)
```

### Layer Responsibilities

| Layer | Package | Responsibility |
|---|---|---|
| **Config** | `config/` | Loads environment variables with sensible defaults |
| **Database** | `db/sqlc/` | Auto-generated type-safe Go code from SQL queries (via SQLC) |
| **Repository** | `internal/repository/` | Wraps SQLC queries, translates `pgx.ErrNoRows` → domain `ErrNotFound` |
| **Service** | `internal/service/` | Business logic — input validation (go-playground/validator), age calculation, DTO mapping |
| **Handler** | `internal/handler/` | HTTP layer — parses requests, calls services, returns JSON with correct status codes |
| **Middleware** | `internal/middleware/` | Cross-cutting: `X-Request-ID` header injection, structured request/response logging |
| **Routes** | `internal/routes/` | Registers all endpoints on the Fiber app |
| **Logger** | `internal/logger/` | Initializes Uber Zap production logger used across all layers |
| **Models** | `internal/models/` | Request/response DTOs, `CalculateAge()` function |

### Why This Architecture?

- **Separation of concerns**: Each layer has a single responsibility. The handler never touches SQL; the repository never knows about HTTP.
- **Testability**: Each layer can be unit-tested independently. The service layer is where business logic lives, making it the easiest place to write meaningful tests.
- **SQLC for type safety**: Instead of writing SQL by hand and scanning rows manually, SQLC generates Go structs and methods from `.sql` files — eliminating an entire class of runtime errors.
- **Dependency injection**: Dependencies flow downward via constructors (`New*` functions), making the wiring explicit in `main.go`.

---

## Tech Stack

| Tool | Purpose |
|---|---|
| [GoFiber](https://gofiber.io/) | High-performance HTTP framework (Express-inspired) |
| [PostgreSQL 16](https://www.postgresql.org/) | Relational database |
| [SQLC](https://sqlc.dev/) | Generates type-safe Go code from SQL |
| [pgx/v5](https://github.com/jackc/pgx) | PostgreSQL driver for Go |
| [Uber Zap](https://github.com/uber-go/zap) | Structured, leveled logging |
| [go-playground/validator](https://github.com/go-playground/validator) | Struct-level input validation |
| [Docker](https://www.docker.com/) | Containerized deployment |

---

## Database Schema

```sql
CREATE TABLE users (
    id   SERIAL PRIMARY KEY,   -- auto-incrementing integer
    name TEXT   NOT NULL,       -- user's name
    dob  DATE   NOT NULL        -- date of birth (age is calculated, not stored)
);
```

**Why not store age?** Age changes every year (or every day on someone's birthday). Storing it would require a cron job or trigger to keep it up-to-date. Instead, we calculate it on the fly in Go — always accurate, zero maintenance.

---

## How Age Calculation Works

```go
func CalculateAge(dob time.Time) int {
    now := time.Now()
    age := now.Year() - dob.Year()

    if now.Month() < dob.Month() ||
        (now.Month() == dob.Month() && now.Day() < dob.Day()) {
        age--
    }
    return age
}
```

1. Start with the year difference.
2. If the birthday hasn't occurred yet this year (month/day comparison), subtract 1.
3. This handles leap years correctly (unlike a `YearDay()` approach which breaks on leap years).

---

## API Endpoints

### Create User
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "dob": "1990-05-10"}'
```
**Response** `201 Created`:
```json
{"id": 1, "name": "Alice", "dob": "1990-05-10"}
```

### Get User by ID
```bash
curl http://localhost:3000/users/1
```
**Response** `200 OK`:
```json
{"id": 1, "name": "Alice", "dob": "1990-05-10", "age": 36}
```

### Update User
```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Updated", "dob": "1991-03-15"}'
```
**Response** `200 OK`:
```json
{"id": 1, "name": "Alice Updated", "dob": "1991-03-15"}
```

### Delete User
```bash
curl -X DELETE http://localhost:3000/users/1
```
**Response** `204 No Content`

### List All Users (Paginated)
```bash
curl "http://localhost:3000/users?page=1&page_size=10"
```
**Response** `200 OK`:
```json
{
  "users": [
    {"id": 1, "name": "Alice", "dob": "1990-05-10", "age": 36}
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

### Health Check
```bash
curl http://localhost:3000/health
```
**Response**: `{"status": "ok"}`

---

## Middleware

### 1. Request ID (`X-Request-ID`)
Every response includes a unique `X-Request-ID` header. If the client sends one, it's preserved; otherwise a UUID v4 is generated. This makes it trivial to trace a request through logs.

### 2. Request Logger
Every request is logged with structured fields:
```json
{"level":"info","msg":"HTTP Request","method":"GET","path":"/users/1","status":200,"duration":"1.234ms","request_id":"abc-123"}
```

### 3. CORS
Cross-Origin Resource Sharing is enabled by default (via Fiber's built-in CORS middleware).

---

## Setup & Run

### Prerequisites
- Go 1.23+
- PostgreSQL 16+ (or Docker)

### Option 1: Docker Compose (Recommended)

```bash
docker-compose up --build
```

This starts both PostgreSQL and the app. The API is available at `http://localhost:3000`.

### Option 2: Local Development

1. **Start PostgreSQL** (e.g., via Docker):
   ```bash
   docker run -d --name userdb \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=userdb \
     -p 5432:5432 \
     postgres:16-alpine
   ```

2. **Configure environment** — edit `.env` or export variables:
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=userdb
   APP_PORT=3000
   ```

3. **Run the app**:
   ```bash
   go run ./cmd/server
   ```

### Running Tests

```bash
go test ./internal/models/ -v
```

---

## Project Structure

```
.
├── cmd/server/main.go           # Entry point — wires everything together
├── config/config.go             # Environment-based configuration
├── db/
│   ├── migrations/              # SQL migration files
│   │   ├── 001_create_users.up.sql
│   │   └── 001_create_users.down.sql
│   ├── queries/users.sql        # SQLC query definitions
│   └── sqlc/                    # SQLC-generated Go code
│       ├── db.go
│       ├── models.go
│       ├── querier.go
│       └── users.sql.go
├── internal/
│   ├── handler/user_handler.go  # HTTP handlers
│   ├── logger/logger.go         # Zap logger setup
│   ├── middleware/
│   │   ├── request_id.go        # X-Request-ID injection
│   │   └── request_logger.go    # Structured request logging
│   ├── models/
│   │   ├── user.go              # DTOs + CalculateAge()
│   │   └── user_test.go         # Unit tests for age calc
│   ├── repository/user_repository.go  # Data access layer
│   ├── routes/routes.go         # Route registration
│   └── service/user_service.go  # Business logic + validation
├── .env                         # Local dev env vars
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Full-stack local deploy
├── sqlc.yaml                    # SQLC configuration
├── go.mod
└── go.sum
```

---

## Error Handling

| Scenario | HTTP Status | Response |
|---|---|---|
| Invalid JSON body | 400 | `{"error": "Invalid request body"}` |
| Validation failure (empty name, bad date) | 400 | `{"error": "...validator details..."}` |
| DOB in the future | 400 | `{"error": "date of birth cannot be in the future"}` |
| User not found | 404 | `{"error": "User not found"}` |
| Database/server error | 500 | `{"error": "Failed to ..."}` |

---

## Interview Talking Points

1. **Why SQLC over an ORM?** SQLC generates code at compile time from raw SQL — you get type safety without the runtime overhead or magic of an ORM. You write SQL you already know, and the generated Go code is plain, reviewable, and fast.

2. **Why calculate age in Go instead of SQL?** Calculating age in the application layer keeps the DB schema simple (just store facts), makes unit testing trivial (pure function), and avoids database-specific date functions that vary between Postgres/MySQL.

3. **Why the repository pattern?** It decouples the service layer from the specific database driver (pgx). The repository translates database-specific errors (`pgx.ErrNoRows`) into domain errors (`ErrNotFound`), so the handler layer never imports database packages.

4. **Why Fiber?** Fiber is built on fasthttp (one of the fastest Go HTTP libraries) and offers an Express.js-like API that's intuitive. It supports middleware chaining, route grouping, and built-in features like CORS out of the box.

5. **How does the middleware work?** The `RequestID` middleware runs first — it generates/preserves a UUID. The `RequestLogger` middleware wraps the handler call, measuring duration from before to after `c.Next()`, then logs a structured JSON entry with the request ID for traceability.

6. **Graceful shutdown** — The server listens for SIGINT/SIGTERM in a goroutine and calls `app.Shutdown()`, which lets in-flight requests finish before closing the process.
