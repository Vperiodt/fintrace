# Fintrace Relationship Visualization

## Overview

Fintrace is a full-stack prototype for analysing relationships between users and transactions in a financial graph. The backend exposes REST APIs to ingest data, detect direct and indirect relationships, and run lightweight analytics (like shortest-path). The frontend renders searchable lists and an interactive graph explorer powered by Cytoscape. Synthetic data generation and container orchestration are provided to support rapid demos with 100k+ transactions.

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 20+ (for frontend dev)
- Docker / Docker Compose (optional but recommended)

### Local Backend

```bash
cd backend
go mod tidy
go test ./...
GRAPH_URI=bolt://localhost:7687 \
GRAPH_USERNAME=neo4j GRAPH_PASSWORD=test \
go run ./cmd/server
```

By default the server listens on `:8080`. Without `GRAPH_URI` it falls back to an in-memory graph (handy for smoke tests, but no persistence).

### Local Frontend

```bash
cd frontend
npm install
npm run dev -- --host
```

The dev server proxies `/users`, `/transactions`, `/relationships`, `/analytics`, and `/export` calls to `http://localhost:8080`.

### Docker Compose (Neo4j + Backend + Frontend)

```bash
docker compose up --build
```

Services:
- `neo4j` (ports `7474`, `7687`)
- `backend` (port `8080`)
- `frontend` (port `5173`, served via nginx)

To generate seed JSON files (10k users / 100k transactions) run:

```bash
docker compose run --rm --profile seed datagen
```

Seed files will land in `./seed-data`.

Then ingest the dataset into Neo4j (the command wires itself up to the running compose stack):

```bash
docker compose --profile seed run --rm ingest
```

Once the CLI reports `ingestion complete`, the `backend` API and React tables can page through the full 100k transactions straight away—filters and sorts operate server-side, so the UI stays responsive even at that scale.

## API Reference (Summary)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/users` | Upsert user + shared attributes |
| `GET`  | `/users` | Paginated list; filters on search, KYC, risk, geography, and email domain |
| `POST` | `/transactions` | Upsert transaction + relationship edges |
| `GET`  | `/transactions` | Paginated list; filters on status, type, channel, amount, time window |
| `GET`  | `/relationships/user/{id}` | User-centric relationships |
| `GET`  | `/relationships/transaction/{id}` | Transaction-centric relationships |
| `GET`  | `/analytics/shortest-path` | Shortest path between two user IDs |
| `GET`  | `/export/users` | Export users (`format=json|csv`) |
| `GET`  | `/export/transactions` | Export transactions (`format=json|csv`) |

All list endpoints support `page`, `pageSize` (<=200), `sortField`, and `sortOrder`.

## Data Generation

The generator produces deterministic data with configurable density of shared attributes:

```bash
cd backend
go run ./cmd/datagen \
  --users 10000 \
  --transactions 100000 \
  --shared-attr-chance 0.35 \
  --output-dir ../data
```

Outputs:
- `data/users.json`
- `data/transactions.json`

## Testing & QA

- Unit tests: `go test ./...`
- Frontend lint/build: `npm run build`
- Manual checklist: see `docs/manual-testing.md`
- Suggested profiling:
  - `go test -run=^$ -bench=. -benchmem ./...`
  - `go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30`

