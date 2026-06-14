// Package storage implements local SQLite history and JSONL event logging for
// the virtual PLC runtime. It must not be imported by pkg/pid, pkg/simulator,
// pkg/plc, or pkg/mqttx. All integration belongs in internal/app.
package storage
