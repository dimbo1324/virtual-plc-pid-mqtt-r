# Demo

The current demo includes the virtual PLC, PID-controlled synthetic processes,
MQTT telemetry and commands. Storage and the web dashboard are not available.

## Without MQTT

```bash
go run ./cmd/vplc --run --config configs/default.json
```

The PLC continues scanning if the configured broker is unavailable. Stop it
with Ctrl+C.

## With Local Mosquitto

```bash
docker compose -f docker/docker-compose.yml up -d
go run ./cmd/vplc --run --config configs/default.json
mosquitto_sub -h localhost -t "vplc/#" -v
```

In another terminal, change the pressure setpoint:

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"demo-1","command":"set_setpoint","loop":"pressure","value":7.5}'
```

Stop the application with Ctrl+C, then stop the broker:

```bash
docker compose -f docker/docker-compose.yml down
```

The included Mosquitto configuration allows anonymous local access for demo
purposes. It is not a production security configuration.
