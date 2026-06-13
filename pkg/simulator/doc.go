// Package simulator implements deterministic synthetic process models for
// virtual PLC simulation. Its first-order model moves a process value (PV)
// toward a target derived from the manipulated value (MV), process base, and
// gain.
//
// Each process owns a seeded random source for reproducible Gaussian noise and
// supports explicit, time-limited disturbance injection. Automatic random
// disturbance scheduling is intentionally reserved for a later stage.
//
// The package is independent from PLC, PID, MQTT, storage, and UI code. It is a
// demonstration model, not a validated representation of a physical plant.
package simulator
