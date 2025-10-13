# Fintrace Relationship Visualization

## Overview

Fintrace is a graph analytics stack: a Go REST backend wiring users and transactions into Neo4j with derived relationships, a React/Cytoscape front-end for interactive exploration, and Dockerized tooling—including a synthetic data generator—to spin up 10k/100k-scale demos quickly.

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

<img width="1861" height="738" alt="image" src="https://github.com/user-attachments/assets/fbd725ef-9ed5-420d-9658-2d8c26ef4247" />
<img width="1831" height="738" alt="image" src="https://github.com/user-attachments/assets/1af54d67-6abf-447c-b4da-a08cb9428176" />
<img width="1831" height="931" alt="image" src="https://github.com/user-attachments/assets/3a3aec41-f775-4da7-a84d-9b6c9fca2c43" />
<img width="1831" height="890" alt="image" src="https://github.com/user-attachments/assets/6b43197b-dfe3-4a1e-b871-3ad8ce5bfb8f" />



