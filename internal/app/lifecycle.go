package app

import (
	"context"
	"fmt"
	"sync"
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

	plcConfig := mapPLCConfig(a.Config)
	runtime, err := plc.NewRuntime(plcConfig)
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

	var background sync.WaitGroup
	if mqttConfig.Enabled {
		if err := mqttClient.Connect(ctx); err != nil {
			a.Logger.Warn("MQTT broker unavailable; PLC runtime continues", "error", err)
			background.Add(1)
			go func() {
				defer background.Done()
				a.retryInitialMQTTConnection(ctx, mqttClient, mqttConfig.ReconnectInterval, mqttConfig.ConnectTimeout)
			}()
		} else {
			a.Logger.Info("MQTT connected", "broker", mqttConfig.BrokerURL, "base_topic", mqttConfig.BaseTopic)
		}
		background.Add(1)
		go func() {
			defer background.Done()
			a.publishMQTTTelemetry(ctx, runtime, mqttClient, plcConfig.PublishInterval)
		}()
	}
	background.Add(1)
	go func() {
		defer background.Done()
		a.forwardRuntimeEvents(ctx, runtime, mqttClient)
	}()

	a.Logger.Info("PLC runtime mode active",
		"device_id", a.Config.App.DeviceID,
		"state", runtime.State(),
		"mqtt_enabled", mqttConfig.Enabled,
	)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	stopErr := runtime.Stop(shutdownCtx)
	background.Wait()
	mqttClient.Disconnect(250)
	if stopErr != nil {
		return fmt.Errorf("stop PLC runtime: %w", stopErr)
	}
	a.Logger.Info("runtime stopped gracefully")
	return nil
}

func (a *App) retryInitialMQTTConnection(ctx context.Context, client *mqttx.Client, interval, connectTimeout time.Duration) {
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
			attemptCtx, cancel := context.WithTimeout(ctx, connectTimeout)
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

func (a *App) publishMQTTTelemetry(ctx context.Context, runtime *plc.Runtime, client *mqttx.Client, publishInterval time.Duration) {
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
		}
	}
}

func (a *App) forwardRuntimeEvents(ctx context.Context, runtime *plc.Runtime, client *mqttx.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-runtime.Events():
			// This is the application-owned fan-out point. Stage 07 can persist
			// events here without adding storage dependencies to pkg/plc.
			if client.IsConnected() && shouldPublishRuntimeEventToMQTT(event) {
				if err := client.PublishEvent(ctx, event); err != nil {
					a.Logger.Warn("publish MQTT event", "error", err)
				}
			}
		}
	}
}

func shouldPublishRuntimeEventToMQTT(event plc.Event) bool {
	if event.Details == nil {
		return true
	}
	// MQTT command responses are already published by pkg/mqttx. Skipping
	// their mirrored runtime events prevents duplicate command outcomes.
	return event.Details["source"] != "mqtt"
}
