// Package pid implements a reusable PID controller for virtual PLC simulation.
// It is independent from application configuration, MQTT, storage, and UI code.
//
// The controller supports auto, manual, hold, and disabled modes. Auto mode
// uses derivative-on-measurement to reduce setpoint derivative kick and
// conditional integration to prevent windup while the output is saturated.
//
// This package is intended for simulation and demonstration. It is not a
// safety-certified controller for real industrial equipment.
package pid
