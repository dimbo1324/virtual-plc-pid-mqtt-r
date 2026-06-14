# Docker Deployment

## Quick start

```bash
docker compose up -d
```

This starts two services:
- **mosquitto** — Eclipse Mosquitto 2 MQTT broker (ports 1883, 9001)
- **vplc-app** — Virtual PLC dashboard (port 8080)

Open the dashboard at http://localhost:8080 after both services are healthy.

## Building the image

```bash
docker build -t vplc-app .
```

The multi-stage `Dockerfile` compiles the Go binary with `CGO_ENABLED=0` so the final image
uses only the Alpine runtime, with no libc dependency.

## Container configuration

The container reads `configs/docker.json` which is baked into the image at build time.
Key differences from the local `configs/default.json`:

| Field | Local default | Docker |
|---|---|---|
| `web.host` | `127.0.0.1` | `0.0.0.0` |
| `mqtt.broker_url` | `tcp://localhost:1883` | `tcp://mosquitto:1883` |
| `storage.sqlite_path` | `data/history.db` | `/app/data/history.db` |
| `storage.events_jsonl_path` | `logs/events.jsonl` | `/app/logs/events.jsonl` |
| `storage.fallback_on_error` | `false` | `true` |
| `storage.fallback_type` | — | `jsonl` |

## Persistent volumes

Two named Docker volumes are declared:

- **vplc-data** — mounted at `/app/data` — SQLite database
- **vplc-logs** — mounted at `/app/logs` — JSONL event log, app log

Data survives container restarts. To wipe all state:

```bash
docker compose down -v
```

## Health checks

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | Liveness probe — always 200 if the process is up |
| `GET /readyz` | Readiness probe — 200 when PLC is running, 503 otherwise |

Both are polled by `docker compose` every 15 s.

## Non-root user

The container runs as user `vplc` (UID/GID created in the image). The `/app` directory
and both volume mount points are pre-chowned to `vplc` during the build.

## Stopping

```bash
docker compose stop        # keeps volumes
docker compose down        # removes containers
docker compose down -v     # removes containers and volumes
```
