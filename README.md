# User API

A RESTful API built in Go for managing users with `name` and `dob` (date of birth). User age is calculated dynamically and never stored in the database.

---

## Tech Stack

- **GoFiber** — HTTP framework
- **PostgreSQL** — Database
- **SQLC** — Type-safe SQL code generation
- **pgx/v5** — PostgreSQL driver
- **Uber Zap** — Structured logging
- **go-playground/validator** — Input validation
- **Docker** — Containerized deployment

---

## Prerequisites

- Go 1.23+
- Docker & Docker Compose (recommended)

---

## Setup & Run

### Using Docker Compose (Recommended)

```bash
go mod tidy
docker-compose up --build
```

The API will be available at `http://localhost:3000`.

### Local Development

1. Start PostgreSQL:
   ```bash
   docker run -d --name userdb \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=userdb \
     -p 5432:5432 \
     postgres:16-alpine
   ```

2. Run the app:
   ```bash
   go mod tidy
   go run ./cmd/server
   ```

### Environment Variables

Configure via `.env` file or system environment:

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `userdb` | Database name |
| `APP_PORT` | `3000` | Application port |

---

## API Endpoints

| Method | Endpoint | Description | Status |
|---|---|---|---|
| `POST` | `/users` | Create a user | `201` |
| `GET` | `/users/:id` | Get user by ID (includes age) | `200` |
| `PUT` | `/users/:id` | Update a user | `200` |
| `DELETE` | `/users/:id` | Delete a user | `204` |
| `GET` | `/users?page=1&page_size=10` | List users (paginated, includes age) | `200` |
| `GET` | `/health` | Health check | `200` |

### Example

```bash
# Create
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "dob": "1990-05-10"}'

# Get (returns calculated age)
curl http://localhost:3000/users/1
```

---

## Project Structure

```
├── cmd/server/main.go          # Entry point
├── config/                     # Environment configuration
├── db/
│   ├── migrations/             # SQL migration files
│   ├── queries/                # SQLC query definitions
│   └── sqlc/                   # SQLC generated code
├── internal/
│   ├── handler/                # HTTP request handlers
│   ├── service/                # Business logic & validation
│   ├── repository/             # Database access layer
│   ├── models/                 # Request/response DTOs
│   ├── middleware/             # Request ID & request logger
│   ├── routes/                 # Route registration
│   └── logger/                 # Zap logger setup
├── Dockerfile
├── docker-compose.yml
└── sqlc.yaml
```

---

## Running Tests

```bash
go test ./internal/models/ -v
```

---

## Features

- Clean layered architecture (Handler → Service → Repository)
- Age calculated dynamically using Go's `time` package
- Input validation with struct tags
- Pagination support on list endpoint
- `X-Request-ID` header on every response
- Request duration logging
- CORS enabled
- Graceful shutdown
- Auto-migration on startup
- Multi-stage Docker build
