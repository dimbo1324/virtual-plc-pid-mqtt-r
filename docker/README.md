# Local MQTT Broker

Docker Compose provides only a local Eclipse Mosquitto broker. The Go
application remains a separately run single binary.

```bash
docker compose -f docker/docker-compose.yml config
docker compose -f docker/docker-compose.yml up -d
docker compose -f docker/docker-compose.yml down
```

Listeners:

- `1883`: MQTT over TCP
- `9001`: MQTT over WebSockets

Anonymous access is enabled solely for local demonstration. Do not expose this
configuration to an untrusted network or treat it as production-ready.
