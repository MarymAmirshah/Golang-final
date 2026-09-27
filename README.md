# GoPower ⚡ — Smart Microgrid Energy Dispatch System

**GoPower** is a backend system for managing a smart microgrid: it tracks power stations and consumer meters, dispatches energy between them under real business constraints (capacity, contracted load, battery storage), exposes everything over a REST API, and streams live grid telemetry and alerts to connected clients over WebSockets.

This project was built as the final capstone assignment for the **Go programming course on [Quera](https://quera.org)**, completed across four incremental phases — from core domain models and persistence, up through a full Gin-based REST API and a concurrent WebSocket hub with graceful shutdown.

## Features

- **Domain models & validation** — `PowerStation` and `ConsumerMeter` entities with field-level validation and sentinel errors (`errors.Is` / `errors.As` friendly).
- **Persistence layer** — GORM-backed repositories over SQLite, with race-safe `Create` (pre-check + unique-constraint fallback), full-record `Update`, and existence-checked `Delete`.
- **Grid dispatch service** — atomic, transactional business logic for:
  - recording generation output from a station,
  - charging battery storage,
  - dispatching energy from a station to a consumer (checked against available energy and the consumer's contracted capacity),
  - producing a grid-wide summary.
- **REST API** (Gin) — full CRUD-style endpoints for stations and consumers, plus generation, dispatch, and grid-summary endpoints, with a spec-accurate HTTP status code contract (`200` / `201` / `400` / `404` / `409` / `422`).
- **Real-time WebSocket hub** — a central hub broadcasts `TELEMETRY` and `ALERT` messages to every connected client, with ping/pong keep-alives and non-blocking, deadlock-free registration/unregistration.
- **Graceful shutdown** — on `SIGINT`/`SIGTERM`, the HTTP server drains in-flight requests before the WebSocket hub is stopped and every connected client is cleanly disconnected.

## Architecture

```
main.go                 → wires everything together, starts the HTTP server, handles graceful shutdown
db/                     → SQLite connection setup (GORM)
models/                 → domain structs, validation, sentinel errors
repository/             → GORM-backed CRUD for stations & consumers
service/                → GridService — the transactional business logic
handlers/               → Gin HTTP handlers (stations, consumers, dispatch, analytics, websocket)
server/                 → router setup, route wiring
ws/                     → WebSocketHub + per-connection Client (read/write pumps)
```

The layering is deliberately one-directional: `handlers` depend on `service`, `service` depends on `repository`, and `repository` depends on `models` — never the other way around.

## API Reference

| Method | Path                     | Description                                   |
|--------|--------------------------|------------------------------------------------|
| GET    | `/health`                | Liveness check                                  |
| GET    | `/ws`                    | Upgrade to a WebSocket connection               |
| POST   | `/api/v1/stations`       | Create a power station                          |
| GET    | `/api/v1/stations`       | List stations (optional `?type=` filter)        |
| GET    | `/api/v1/stations/:id`   | Get a station by ID                             |
| POST   | `/api/v1/consumers`      | Create a consumer meter                         |
| GET    | `/api/v1/consumers`      | List consumers                                  |
| GET    | `/api/v1/consumers/:id`  | Get a consumer by ID                            |
| POST   | `/api/v1/generation`     | Record generation output for a station          |
| POST   | `/api/v1/dispatch`       | Dispatch energy from a station to a consumer     |
| GET    | `/api/v1/grid/summary`   | Grid-wide summary (generation, load, status)    |

### WebSocket messages

Connected clients receive JSON messages of two kinds:

```json
{ "type": "TELEMETRY", "summary": { ... GridSummary ... }, "timestamp": "..." }
{ "type": "ALERT", "severity": "WARNING", "message": "High grid frequency detected", "timestamp": "..." }
```

## Getting Started

### Prerequisites
- Go 1.24+

### Run locally

```bash
go mod download
go run main.go
```

The server listens on `:8080`. Try it:

```bash
curl http://localhost:8080/health
```

### Run the test suite

```bash
go test ./...
# with the race detector (recommended for the ws/ and service/ packages)
go test -race ./...
```

### Graceful shutdown

Press `Ctrl+C` (or send `SIGTERM`) — the server stops accepting new connections, drains in-flight HTTP requests, then closes all active WebSocket connections before exiting.

## Tech Stack

- [Gin](https://github.com/gin-gonic/gin) — HTTP web framework
- [GORM](https://gorm.io/) + SQLite driver — persistence
- [gorilla/websocket](https://github.com/gorilla/websocket) — WebSocket connections

## About

Built and delivered in four phases as a Quera.org Go course final project: domain models & repositories → grid dispatch service → REST API → real-time WebSocket layer with graceful shutdown.
