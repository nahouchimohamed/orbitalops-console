# OrbitalOps Console

Production-shaped **Spacecraft Control Subsystem (SCS) lite** for a small SAR constellation.

Operators can:

1. Load a seeded satellite fleet (TLE-backed)
2. Visualize predicted **ground tracks** on a map
3. Compute ground-station **AOS/LOS** pass windows
4. Submit an observation request (AOI) and review a **draft imaging plan**
5. Export a mock **command sequence** JSON for the human-in-the-loop workflow

## Architecture

```
React UI  →  api-gateway (REST BFF)  →  planning-service (REST + Postgres)
                                   ↘  orbit-service (gRPC / SGP4)
```

See [docs/architecture.md](docs/architecture.md).

## Stack

| Layer | Tech |
|-------|------|
| Frontend | React, TypeScript, Vite, Deck.gl, MapLibre |
| Gateway | Go, chi, REST |
| Planning | Go, chi, Postgres, golang-migrate |
| Orbit | Go, gRPC, SGP4 (`go-satellite`) |
| Runtime | Docker Compose |

## Quick start

```bash
docker compose up --build
```

- UI: http://localhost:5173  
- API gateway: http://localhost:8080  
- Planning: http://localhost:8081  
- Orbit gRPC: localhost:50051  
- Postgres: localhost:5432  

```bash
# run migrations (also runs on planning-service startup)
docker compose run --rm migrate

# tests
make test
```

## Demo flow

1. Open the UI → fleet table loads seeded StriX-like satellites  
2. Select a satellite → ground track renders on the map  
3. Pick a ground station → upcoming AOS/LOS windows appear  
4. Draw / submit an AOI observation request → draft plan + timeline  
5. Export command sequence JSON  

## License

MIT
