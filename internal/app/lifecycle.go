package app

import (
	"context"
	"fmt"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/mqttx"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

const shutdownTimeout = 5 * time.Second

// RunRuntime starts the optional long-running PLC and MQTT services.
func (a *App) RunRuntime(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	runtime, err := plc.NewRuntime(mapPLCConfig(a.Config))
	if err != nil {
		return fmt.Errorf("create PLC runtime: %w", err)
	}
	if a.Config.App.AutoStart {
		if err := runtime.Start(ctx); err != nil {
			return fmt.Errorf("start PLC runtime: %w", err)
		}
	}

	mqttConfig := mapMQTTConfig(a.Config)
	mqttClient, err := mqttx.New(mqttConfig, func(_ context.Context, command plc.Command) (plc.Event, error) {
		return runtime.ApplyCommand(command)
	})
	if err != nil {
		_ = runtime.Stop(context.Background())
		return fmt.Errorf("create MQTT client: %w", err)
	}

	if mqttConfig.Enabled {
		if err := mqttClient.Connect(ctx); err != nil {
			a.Logger.Warn("MQTT broker unavailable; PLC runtime continues", "error", err)
			go a.retryInitialMQTTConnection(ctx, mqttClient, mqttConfig.ReconnectInterval)
		} else {
			a.Logger.Info("MQTT connected", "broker", mqttConfig.BrokerURL, "base_topic", mqttConfig.BaseTopic)
		}
		go a.bridgeMQTT(ctx, runtime, mqttClient, mapPLCConfig(a.Config).PublishInterval)
	}

	a.Logger.Info("PLC runtime mode active",
		"device_id", a.Config.App.DeviceID,
		"state", runtime.State(),
		"mqtt_enabled", mqttConfig.Enabled,
	)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := runtime.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("stop PLC runtime: %w", err)
	}
	mqttClient.Disconnect(250)
	a.Logger.Info("runtime stopped gracefully")
	return nil
}

func (a *App) retryInitialMQTTConnection(ctx context.Context, client *mqttx.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if client.IsConnected() {
				return
			}
			attemptCtx, cancel := context.WithTimeout(ctx, mapMQTTConfig(a.Config).ConnectTimeout)
			err := client.Connect(attemptCtx)
			cancel()
			if err == nil {
				a.Logger.Info("MQTT connected after retry")
				return
			}
			a.Logger.Warn("MQTT reconnect attempt failed", "error", err)
		}
	}
}

func (a *App) bridgeMQTT(ctx context.Context, runtime *plc.Runtime, client *mqttx.Client, publishInterval time.Duration) {
	ticker := time.NewTicker(publishInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if client.IsConnected() {
				if err := client.PublishSnapshot(ctx, runtime.Snapshot()); err != nil {
					a.Logger.Warn("publish MQTT telemetry", "error", err)
				}
			}
		case event := <-runtime.Events():
			if client.IsConnected() {
				if err := client.PublishEvent(ctx, event); err != nil {
					a.Logger.Warn("publish MQTT event", "error", err)
				}
			}
		}
	}
}
