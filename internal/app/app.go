package app

import (
	"context"
	"log/slog"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
)

// App contains configuration and logging used to assemble runtime services.
type App struct {
	Config config.Config
	Logger *slog.Logger
}

// New creates an application foundation instance.
func New(cfg config.Config, logger *slog.Logger) *App {
	return &App{Config: cfg, Logger: logger}
}

// Run preserves the short foundation startup used by CLI smoke checks.
func (a *App) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.Logger.Info("Virtual PLC PID MQTT Runtime starting...",
		"device_id", a.Config.App.DeviceID,
	)
	a.Logger.Info("runtime foundation initialized")

	return nil
}
