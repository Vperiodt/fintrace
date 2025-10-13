# Fintrace Relationship Visualization

## Overview

Fintrace is a full-stack prototype for analysing relationships between users and transactions in a financial graph. The backend exposes REST APIs to ingest data, detect direct and indirect relationships. The frontend renders searchable lists and an interactive graph explorer powered by Cytoscape. Synthetic data generation and container orchestration are provided to support rapid demos with 100k+ transactions.

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 20+ (for frontend dev)
- Docker / Docker Compose (optional but recommended)


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
docker compose --profile seed run --rm datagen
```

Seed files will land in `./seed-data`.

Then ingest the dataset into Neo4j. The loader retries transient deadlocks automatically, but for a guaranteed smooth import start with a single worker:

```bash
docker compose --profile seed run --rm ingest --dataset-dir /seed-data --workers 1
```

### Quick demo vs. full dataset

If you want a tiny dataset that highlights every relationship type without waiting for the full import, load the curated demo files instead:

```bash
# 6 users, 12 transactions (loads in seconds)
docker compose --profile demo run --rm ingest-demo

# Full 10k / 100k dataset
docker compose --profile seed run --rm datagen
docker compose --profile seed run --rm ingest --dataset-dir /seed-data --workers 1
```

### Testing the API quickly

With the stack running (`docker compose up backend frontend`), you can smoke test the REST endpoints with `curl`:

```bash
# Users list
curl -s "http://localhost:8080/users?page=1&pageSize=5" | jq

# Transactions filtered by status
curl -s "http://localhost:8080/transactions?page=1&pageSize=5&status=COMPLETED" | jq

# User relationships
curl -s "http://localhost:8080/relationships/user/USR-DEMO-1" | jq

# Transaction relationships
curl -s "http://localhost:8080/relationships/transaction/TX-DEMO-1" | jq
```

> You can create users and transactions via the dashboard forms or directly with the API (e.g. curl/Postman) if you prefer scripting.

## API Reference (Summary)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/users` | Upsert user + shared attributes |
| `GET`  | `/users` | Paginated list; filters on search, KYC, risk, geography, and email domain |
| `POST` | `/transactions` | Upsert transaction + relationship edges |
| `GET`  | `/transactions` | Paginated list; filters on status, type, channel, amount, time window |
| `GET`  | `/relationships/user/{id}` | User-centric relationships |
| `GET`  | `/relationships/transaction/{id}` | Transaction-centric relationships |

All list endpoints support `page`, `pageSize` (<=200), `sortField`, and `sortOrder`.
