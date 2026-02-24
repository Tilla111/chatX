# ChatX

<p align="center">
  <img src="./docs/images/chatx-cover.svg" alt="ChatX cover banner" />
</p>

<p align="center">
  Real-time chat platform built with Go, JWT auth, WebSocket events, and PostgreSQL.
</p>

---

## Table of Contents

1. [About](#about)
2. [Visual Overview](#visual-overview)
3. [Core Features](#core-features)
4. [Tech Stack](#tech-stack)
5. [Run Locally](#run-locally)
6. [Configuration](#configuration)
7. [API Conventions](#api-conventions)
8. [Endpoint Map](#endpoint-map)
9. [WebSocket Events](#websocket-events)
10. [Data Model](#data-model)
11. [Project Structure](#project-structure)
12. [Make Commands](#make-commands)
13. [Swagger Docs](#swagger-docs)
14. [Troubleshooting](#troubleshooting)

---

## About

`ChatX` is a production-style chat backend and web workspace with:

- JWT-based authentication.
- Email activation flow (Mailtrap SMTP).
- Private and group chats.
- Message CRUD and read tracking.
- Member management with roles (`owner`, `admin`, `member`).
- Real-time WebSocket broadcasts.
- Swagger API documentation.

API base URL:

```text
http://localhost:8080/api/v1
```

---

## Visual Overview

### Workspace UI (schematic)

![ChatX UI overview](./docs/images/chatx-ui-overview.svg)

### Architecture

![ChatX architecture](./docs/images/chatx-architecture.svg)

### Auth flow

![ChatX auth flow](./docs/images/chatx-auth-flow.svg)

### Realtime flow

![ChatX realtime flow](./docs/images/chatx-realtime-flow.svg)

---

## Core Features

- `Register -> Email activation -> Login` flow.
- Only active users (`is_active=true`) can log in.
- `Authorization: Bearer <token>` middleware validation.
- `GET /users` supports pagination (`limit`, `offset`) and `search` (max 10 chars).
- Private chat creation (`POST /chats`) with reuse of an existing private chat.
- Group chat creation/update and member add/remove.
- Message create/read/update/delete and mark-as-read.
- WebSocket event types:
  - `new_message`
  - `message_updated`
  - `message_deleted`
  - `messages_read`
  - `member_added`
  - `chat_deleted`

---

## Tech Stack

- Go `1.25.5`
- Chi router (`github.com/go-chi/chi/v5`)
- PostgreSQL `16.x`
- JWT (`github.com/golang-jwt/jwt/v5`)
- Gorilla WebSocket (`github.com/gorilla/websocket`)
- Validator (`github.com/go-playground/validator/v10`)
- Swagger (`swaggo/swag`, `http-swagger`)
- Mailtrap SMTP

---

## Run Locally

### 1) Requirements

- Go `1.25+`
- Docker + Docker Compose
- `migrate` CLI (`golang-migrate/migrate`)
- Optional: `swag` CLI

### 2) Environment file

Create or update `.env`:

```env
APP_ENV=dev
PORT=8080
API_URL=localhost:8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=admin
DB_NAME=chat_db
DB_MAX_IDLE_CONNS=6
DB_MAX_OPEN_CONNS=9
DB_MAX_IDLE_TIME=15m

MAILTRAP_HOST=sandbox.smtp.mailtrap.io
MAILTRAP_PORT=2525
MAILTRAP_USERNAME=your_mailtrap_username
MAILTRAP_PASSWORD=your_mailtrap_password
FROM_EMAIL=noreply@example.com

JWT_SECRET_KEY=your-secret-key
JWT_AUDIENCE=chatX Audience
JWT_ISSUER=chatX Issuer
```

### 3) Start PostgreSQL

```bash
docker compose up -d
```

### 4) Apply migrations

```bash
make migrate-up
```

Or manually:

```bash
migrate -path=./cmd/migrate/migrations -database="postgres://admin:admin@localhost:5432/chat_db?sslmode=disable" up
```

### 5) Run API server

```bash
go run ./cmd/api
```

### 6) Open URLs

- Web workspace: `http://localhost:8080/`
- Health endpoint: `http://localhost:8080/api/v1/health`
- Swagger UI: `http://localhost:8080/api/v1/swagger/index.html`

---

## Configuration

Config file is selected by `APP_ENV`:

- `APP_ENV=dev` -> `internal/env/config.dev.yaml`
- `APP_ENV=prod` -> `internal/env/config.prod.yaml`

Environment variables override YAML values.

| Variable | Purpose | Default/Fallback |
| --- | --- | --- |
| `APP_ENV` | Runtime environment | `dev` |
| `PORT` | HTTP server port | `8080` |
| `API_URL` | Swagger host and activation URL base | `localhost:8080` |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Database connection values | `config.*.yaml` |
| `JWT_SECRET_KEY` | JWT signing key | value from `config.dev.yaml` |
| `JWT_AUDIENCE`, `JWT_ISSUER` | JWT claims validation values | value from `config.dev.yaml` |
| `MAILTRAP_*`, `FROM_EMAIL` | SMTP activation email settings | value from `config.dev.yaml` |

---

## API Conventions

- Base path: `/api/v1`
- Success response envelope:

```json
{
  "data": {}
}
```

- Error response envelope:

```json
{
  "error": "error message"
}
```

- Auth header:

```http
Authorization: Bearer <JWT_TOKEN>
```

- JSON parsing rules:
  - Unknown fields are rejected (`DisallowUnknownFields`).
  - Maximum request body size is `1MB`.

- WebSocket auth:
  - Standard Bearer token through HTTP auth middleware.
  - Browser fallback is supported:
    - `/api/v1/ws?token=<JWT_TOKEN>`

---

## Endpoint Map

### System

| Method | Endpoint | Auth | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | No | API health check |

### Authentication / Users

| Method | Endpoint | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/users/authentication` | No | Register + send activation email |
| `POST` | `/users/authentication/token` | No | Login and get JWT |
| `GET` | `/users/activate/{token}` | No | Activation endpoint (link use case) |
| `PUT` | `/users/activate/{token}` | No | Activation endpoint (API style) |
| `GET` | `/users` | Yes | User list with pagination/search |

### Chats / Groups

| Method | Endpoint | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/chats` | Yes | Create private chat |
| `GET` | `/chats` | Yes | List current user's chats |
| `DELETE` | `/chats/{chat_id}` | Yes | Delete chat |
| `GET` | `/chats/{chat_id}/messages` | Yes | Get chat messages |
| `POST` | `/groups` | Yes | Create group chat |
| `PATCH` | `/groups/{chat_id}` | Yes | Update group metadata |

### Members

| Method | Endpoint | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/groups/{chat_id}/members` | Yes | Add member (owner/admin) |
| `GET` | `/groups/{chat_id}/members` | Yes | List group members |
| `DELETE` | `/groups/{chat_id}/{user_id}/member` | Yes | Remove member |

### Messages

| Method | Endpoint | Auth | Description |
| --- | --- | --- | --- |
| `POST` | `/messages` | Yes | Send message |
| `PATCH` | `/messages/{id}` | Yes | Update message (sender only) |
| `DELETE` | `/messages/{id}` | Yes | Delete message (sender only) |
| `PATCH` | `/messages/chats/{chat_id}/read` | Yes | Mark chat messages as read |

### Quick `curl` examples

Register:

```bash
curl -X POST http://localhost:8080/api/v1/users/authentication \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"ali\",\"email\":\"ali@example.com\",\"password\":\"secret123\"}"
```

Login:

```bash
curl -X POST http://localhost:8080/api/v1/users/authentication/token \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"ali@example.com\",\"password\":\"secret123\"}"
```

List chats:

```bash
curl http://localhost:8080/api/v1/chats \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## WebSocket Events

### Browser connection example

```js
const token = "<JWT_TOKEN>";
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws?token=${token}`);

ws.onmessage = (event) => {
  const payload = JSON.parse(event.data);
  console.log(payload.type, payload);
};
```

### Event contract

| `type` | Main fields |
| --- | --- |
| `new_message` | `chat_id`, `chat_name`, `sender_id`, `sender_name`, `content`, `created_at` |
| `message_updated` | `chat_id`, `message_id`, `message_text` |
| `message_deleted` | `chat_id`, `message_id` |
| `messages_read` | `chat_id`, `reader_id` |
| `member_added` | `chat_id`, `user_id`, `username`, `added_by_id`, `added_by_name` |
| `chat_deleted` | `chat_id`, `deleted_by_id`, `deleted_by_name` |

---

## Data Model

![ChatX data model](./docs/images/chatx-data-model.svg)

High-level entities:

- `users`
- `user_invitations`
- `chats`
- `chat_members`
- `group_info`
- `messages`
- `message_reads`

---

## Project Structure

```text
chatX/
├─ cmd/
│  ├─ api/                  # HTTP server, middleware, handlers
│  └─ migrate/migrations/   # SQL migrations
├─ internal/
│  ├─ auth/                 # JWT auth service
│  ├─ db/                   # PostgreSQL connection
│  ├─ env/                  # YAML + ENV config loader
│  ├─ mailer/               # Mailtrap client + templates
│  ├─ store/                # Repository layer
│  ├─ usecase/              # Business logic
│  └─ ws/                   # WebSocket hub/client
├─ docs/                    # Swagger output + images
├─ web/                     # Frontend workspace (HTML/CSS/JS)
├─ docker-compose.yml       # Local PostgreSQL service
└─ Makefile                 # Build/run/migrate/docs commands
```

---

## Make Commands

| Command | Purpose |
| --- | --- |
| `make build-api` | Build API binary (`bin/chatx-api`) |
| `make run-api` | Run API server |
| `make migration <name>` | Create migration files |
| `make migrate-up` | Apply migrations |
| `make migrate-down <n>` | Roll back migrations |
| `make docker-up` | Start Docker services |
| `make docker-down` | Stop Docker services |
| `make docker-logs` | Show Docker logs |
| `make gen-docs` | Regenerate Swagger docs |

---

## Swagger Docs

- Swagger UI: `/api/v1/swagger/index.html`
- Swagger source files: `docs/swagger.yaml`, `docs/swagger.json`

When annotations change:

```bash
make gen-docs
```

---

## Troubleshooting

- `401 invalid or expired token`
  - Re-login to get a fresh token.
  - Verify `JWT_SECRET_KEY`, `JWT_AUDIENCE`, and `JWT_ISSUER`.

- `invalid Authorization header format`
  - Must be exactly:
    - `Authorization: Bearer <token>`

- `dial tcp ... connection refused`
  - Verify PostgreSQL is running:
    - `docker compose ps`

- Activation email is not delivered
  - Check `MAILTRAP_*` and `FROM_EMAIL`.

- WebSocket connection fails
  - Verify URL: `ws://localhost:8080/api/v1/ws?token=<JWT>`
  - Local origin should match allowed values (`localhost:8080` or `127.0.0.1:8080`).

---

## License

If there is no dedicated `LICENSE` file yet, choose one with the repository owner (for example, MIT).
