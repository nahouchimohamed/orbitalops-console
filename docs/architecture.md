# Architecture

OrbitalOps Console is a local, production-shaped monorepo that mirrors a Spacecraft Control Subsystem (SCS) operator portal.

## Services

| Service | Protocol | Responsibility |
|---------|----------|----------------|
| `orbit-service` | gRPC `:50051` + HTTP health `:8082` | TLE propagation, ground-track samples, AOS/LOS pass search |
| `planning-service` | REST `:8081` | Fleet/stations CRUD reads, observation requests, draft imaging plans, command sequence JSON |
| `api-gateway` | REST `:8080` | BFF, CORS, system status aggregation, reverse-proxy to planning |
| `frontend` | HTTP `:5173` / `:80` | Operator UI (fleet, Deck.gl tracks, passes, planning timeline) |
| `postgres` | `:5432` | Durable planning state |

## Clean architecture (Go)

Each Go service follows:

```
cmd/server          → composition root
internal/domain     → entities + ports
internal/usecase    → application services
internal/adapter/*  → HTTP/gRPC/Postgres/clients
```

Orbit propagation uses a Keplerian propagator derived from TLE mean elements (SGP4-class workflow). Pass search evaluates elevation vs ground stations over a time window.

## Data flow

1. UI loads satellites + stations via gateway → planning-service / Postgres  
2. Ground track: planning asks orbit-service gRPC `Propagate`  
3. Passes: planning fans out concurrent `FindPasses` per satellite  
4. Observation request: planning stores AOI, scores nearest overflights, writes draft plans + mock command sequence  

## Reliability practices

- Health (`/healthz`) and readiness (`/readyz`) on services  
- Request IDs via chi middleware / gateway headers  
- Context timeouts on outbound gRPC and HTTP  
- Goroutine fan-out for constellation-scale pass search  
- SQL migrations with `golang-migrate`  
- CI: Go tests + frontend Vitest + Docker image builds  

## Local runtime

`docker compose up --build` starts the full stack. Migrations run as a one-shot Compose service before planning starts.
