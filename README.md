# Task Management API

A REST API for managing projects and tasks, built with Go.

## Tech Stack

- **Go** — language
- **Chi** — HTTP router
- **PostgreSQL** — database
- **sqlc** — type-safe SQL query generation
- **goose** — database migrations
- **Google OAuth2** — authentication
- **JWT** — session tokens

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL
- goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- sqlc: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### Setup

1. Clone the repository

```bash
   git clone https://github.com/ochochecharles/task-management-api
   cd task-management-api
```

2. Install dependencies

```bash
   go mod download
```

3. Create a `.env` file in the project root

```env
   DB_URL=postgres://user:password@localhost:5432/task_management?sslmode=disable
   PORT=8080
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
   JWT_SECRET=your-jwt-secret
```

4. Run database migrations

```bash
   goose -dir internal/db/migrations postgres "your-db-url" up
```

5. Start the server

```bash
   go run main.go
```

## Authentication

This API uses Google OAuth2. To authenticate:

1. Visit `GET /auth/google` — redirects to Google login
2. After login, Google redirects to `/auth/google/callback`
3. The response contains a JWT token
4. Include the token in all subsequent requests as `Authorization: Bearer <token>`

## API Endpoints

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects` | Create a project |
| GET | `/projects` | List your projects |
| GET | `/projects/{id}` | Get a project |
| PUT | `/projects/{id}` | Update a project |
| DELETE | `/projects/{id}` | Delete a project |
| POST | `/projects/{id}/members` | Add a member |
| DELETE | `/projects/{id}/members/{userID}` | Remove a member |

### Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/projects/{id}/tasks` | Create a task |
| GET | `/projects/{id}/tasks` | List tasks in a project |
| GET | `/projects/{id}/tasks/{taskID}` | Get a task |
| PUT | `/projects/{id}/tasks/{taskID}` | Update a task |
| DELETE | `/projects/{id}/tasks/{taskID}` | Delete a task |

## Running Tests

```bash
go test ./internal/handler/... -v
```

Tests run against a separate `task_management_test` database. Make sure it exists and migrations have been applied before running tests.