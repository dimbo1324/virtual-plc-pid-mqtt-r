package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/storage"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/web"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/input"
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

	// --- Storage initialization ---
	var rec *storage.Recorder
	var store *storage.Store
	var jsonlWriter *storage.JSONLWriter
	storageMode := "disabled"

	storageCfg := mapStorageConfig(a.Config)
	if storageCfg.Enabled {
		var openErr error
		store, openErr = storage.Open(ctx, storageCfg)
		if openErr != nil {
			if !storageCfg.FallbackOnError {
				return fmt.Errorf("open storage: %w", openErr)
			}
			a.Logger.Warn("storage unavailable; entering degraded mode",
				"error", openErr, "fallback", storageCfg.FallbackType)
			storageMode = "degraded"
			if storageCfg.FallbackType == "jsonl" {
				jw, jwErr := storage.NewJSONLWriter(storageCfg.EventsJSONLPath)
				if jwErr != nil {
					a.Logger.Warn("jsonl fallback failed; events will not be persisted", "error", jwErr)
				} else {
					jsonlWriter = jw
				}
			}
		} else {
			var err error
			jsonlWriter, err = storage.NewJSONLWriter(storageCfg.EventsJSONLPath)
			if err != nil {
				_ = store.Close()
				return fmt.Errorf("open jsonl writer: %w", err)
			}
			rec = storage.NewRecorder(store, jsonlWriter, storageCfg.WriteQueueSize, a.Logger)
			rec.Start(ctx)
			storageMode = "ok"
			a.Logger.Info("storage initialized", "path", storageCfg.SQLitePath)
		}
	}

	// --- PLC runtime ---
	plcConfig := mapPLCConfig(a.Config)
	runtime, err := plc.NewRuntime(plcConfig)
	if err != nil {
		a.closeStorage(store, jsonlWriter, rec)
		return fmt.Errorf("create PLC runtime: %w", err)
	}
	if a.Config.App.AutoStart {
		if err := runtime.Start(ctx); err != nil {
			a.closeStorage(store, jsonlWriter, rec)
			return fmt.Errorf("start PLC runtime: %w", err)
		}
	}

	// --- MQTT client with command recording wrapper ---
	mqttConfig := mapMQTTConfig(a.Config)
	commandHandler := a.buildCommandHandler(runtime, rec)
	mqttClient, err := mqttx.New(mqttConfig, commandHandler)
	if err != nil {
		_ = runtime.Stop(context.Background())
		a.closeStorage(store, jsonlWriter, rec)
		return fmt.Errorf("create MQTT client: %w", err)
	}

	// --- Live storage mode (readable by /api/status at any point in time) ---
	var storageModeAtom atomic.Value
	storageModeAtom.Store(storageMode)
	storageStatusFn := func() string { return storageModeAtom.Load().(string) }
	setStorageMode := func(m string) { storageModeAtom.Store(m) }

	// --- Web dashboard fan-out channels ---
	webCfg := mapWebConfig(a.Config)
	var webEventsCh chan plc.Event
	var webSnapshotsCh chan plc.Snapshot
	if webCfg.Enabled {
		webEventsCh = make(chan plc.Event, 64)
		webSnapshotsCh = make(chan plc.Snapshot, 16)
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

	// Forward runtime events to MQTT, storage, and web SSE.
	// In degraded mode (rec==nil but jsonlWriter!=nil) events go directly to JSONL;
	// persistent write errors update storageMode to "failed" via setStorageMode.
	background.Add(1)
	go func() {
		defer background.Done()
		a.forwardRuntimeEvents(ctx, runtime, mqttClient, rec, jsonlWriter, webEventsCh, setStorageMode)
	}()

	// Consume snapshot stream for storage and web SSE.
	if rec != nil || webSnapshotsCh != nil {
		background.Add(1)
		go func() {
			defer background.Done()
			a.consumeSnapshots(ctx, runtime, rec, webSnapshotsCh)
		}()
	}

	// --- InputProvider hook.
	// Wire a real pkg/input.Provider here to feed external PV values (OPC-UA, Modbus, REST).
	// Example:
	//   provider := mypackage.NewOPCUAProvider(...)
	//   background.Add(1)
	//   go func() { defer background.Done(); a.runInputProvider(ctx, runtime, provider, plcConfig.ScanInterval) }()

	// --- Web dashboard server ---
	if webCfg.Enabled {
		webServer := web.NewServer(webCfg, web.Deps{
			Runtime:         runtime,
			Store:           store,
			CommandHandler:  web.CommandHandler(commandHandler),
			EventsCh:        webEventsCh,
			SnapshotsCh:     webSnapshotsCh,
			StorageStatusFn: storageStatusFn,
		}, a.Logger)
		background.Add(1)
		go func() {
			defer background.Done()
			if err := webServer.Start(ctx); err != nil {
				a.Logger.Error("web server stopped", "error", err)
			}
		}()
	}

	a.Logger.Info("PLC runtime mode active",
		"device_id", a.Config.App.DeviceID,
		"state", runtime.State(),
		"mqtt_enabled", mqttConfig.Enabled,
		"storage_enabled", storageCfg.Enabled,
		"web_enabled", webCfg.Enabled,
	)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	stopErr := runtime.Stop(shutdownCtx)
	background.Wait()
	mqttClient.Disconnect(250)

	// Drain storage recorder, then close resources.
	a.closeStorage(store, jsonlWriter, rec)

	if stopErr != nil {
		return fmt.Errorf("stop PLC runtime: %w", stopErr)
	}
	a.Logger.Info("runtime stopped gracefully")
	return nil
}

// buildCommandHandler wraps runtime.ApplyCommand to record commands when
// storage is enabled. It does not modify pkg/plc or pkg/mqttx.
func (a *App) buildCommandHandler(runtime *plc.Runtime, rec *storage.Recorder) mqttx.CommandHandler {
	return func(cmdCtx context.Context, command plc.Command) (plc.Event, error) {
		event, err := runtime.ApplyCommand(command)

		if rec == nil {
			return event, err
		}

		status := "applied"
		errMsg := ""
		if err != nil {
			status = "rejected"
			errMsg = err.Error()
		}

		payloadJSON := buildCommandPayloadJSON(command)
		cmdRecord := storage.CommandRecord{
			Timestamp:    command.ReceivedAt,
			CommandID:    command.CommandID,
			Source:       command.Source,
			CommandType:  string(command.Command),
			LoopName:     command.Loop,
			PayloadJSON:  payloadJSON,
			Status:       status,
			ErrorMessage: errMsg,
		}
		if cmdRecord.Timestamp.IsZero() {
			cmdRecord.Timestamp = time.Now().UTC()
		}
		rec.RecordCommand(cmdRecord)

		// Record PID tuning changes derived from the resulting event.
		if err == nil && event.Type == plc.EventPIDTuningChanged {
			a.recordPIDChange(rec, command, event)
		}

		return event, err
	}
}

func buildCommandPayloadJSON(command plc.Command) string {
	payload := map[string]any{
		"command": command.Command,
	}
	if command.Loop != "" {
		payload["loop"] = command.Loop
	}
	if command.CommandID != "" {
		payload["command_id"] = command.CommandID
	}
	if command.Value != nil {
		payload["value"] = *command.Value
	}
	if command.Kp != nil {
		payload["kp"] = *command.Kp
	}
	if command.Ki != nil {
		payload["ki"] = *command.Ki
	}
	if command.Kd != nil {
		payload["kd"] = *command.Kd
	}
	if command.Mode != "" {
		payload["mode"] = command.Mode
	}
	if command.ManualOutput != nil {
		payload["manual_output"] = *command.ManualOutput
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (a *App) recordPIDChange(rec *storage.Recorder, command plc.Command, event plc.Event) {
	change := storage.PIDChangeRecord{
		Timestamp: event.Timestamp,
		LoopName:  command.Loop,
		Source:    command.Source,
		CommandID: command.CommandID,
	}
	// New gains from the command (validated before reaching here).
	if command.Kp != nil {
		change.NewKp = *command.Kp
	}
	if command.Ki != nil {
		change.NewKi = *command.Ki
	}
	if command.Kd != nil {
		change.NewKd = *command.Kd
	}
	// Old gains are not exposed by the current PLC event — left as nil.
	rec.RecordPIDChange(change)
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

func (a *App) forwardRuntimeEvents(ctx context.Context, runtime *plc.Runtime, client *mqttx.Client, rec *storage.Recorder, jw *storage.JSONLWriter, webCh chan<- plc.Event, setMode func(string)) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-runtime.Events():
			storageEvent := plcEventToStorageRecord(event)
			if rec != nil {
				rec.RecordEvent(storageEvent)
			} else if jw != nil {
				// Degraded mode: write events directly to JSONL without a recorder.
				if err := jw.WriteEvent(ctx, storageEvent); err != nil {
					a.Logger.Warn("degraded jsonl write", "error", err)
					if setMode != nil {
						setMode("failed")
					}
				}
			}
			if webCh != nil {
				select {
				case webCh <- event:
				default:
				}
			}
			if client.IsConnected() && shouldPublishRuntimeEventToMQTT(event) {
				if err := client.PublishEvent(ctx, event); err != nil {
					a.Logger.Warn("publish MQTT event", "error", err)
				}
			}
		}
	}
}

// consumeSnapshots reads from the runtime snapshot channel and fans out to
// the storage recorder and web SSE channel (both optional).
func (a *App) consumeSnapshots(ctx context.Context, runtime *plc.Runtime, rec *storage.Recorder, webCh chan<- plc.Snapshot) {
	for {
		select {
		case <-ctx.Done():
			return
		case snap := <-runtime.Snapshots():
			if rec != nil {
				if !rec.RecordSnapshot(snap) {
					a.Logger.Warn("storage snapshot queue full; sample dropped")
				}
			}
			if webCh != nil {
				select {
				case webCh <- snap:
				default:
				}
			}
		}
	}
}

func plcEventToStorageRecord(event plc.Event) storage.EventRecord {
	return storage.EventRecord{
		Timestamp: event.Timestamp,
		Level:     event.Level,
		Type:      event.Type,
		Message:   event.Message,
		Details:   event.Details,
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

// runInputProvider polls provider at interval and injects good-quality tag values
// into the runtime as external PV overrides. Exits when ctx is cancelled.
func (a *App) runInputProvider(ctx context.Context, runtime *plc.Runtime, provider input.Provider, interval time.Duration) {
	defer func() { _ = provider.Close() }()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tags, err := provider.Read(ctx)
			if err != nil {
				a.Logger.Warn("input provider read error", "provider", provider.Name(), "error", err)
				continue
			}
			for _, tv := range tags {
				switch tv.Quality {
				case input.QualityGood:
					runtime.InjectPV(tv.Name, tv.Value)
				case input.QualityBad:
					// Bad quality: clear stale injected PV so simulator resumes.
					runtime.ClearPV(tv.Name)
				}
				// QualityUncertain: retain last good value.
			}
		}
	}
}

func (a *App) closeStorage(store *storage.Store, jsonlWriter *storage.JSONLWriter, rec *storage.Recorder) {
	if rec != nil {
		_ = rec.Stop(context.Background())
	}
	if jsonlWriter != nil {
		if err := jsonlWriter.Close(); err != nil {
			a.Logger.Warn("close jsonl writer", "error", err)
		}
	}
	if store != nil {
		if err := store.Close(); err != nil {
			a.Logger.Warn("close storage", "error", err)
		}
	}
}
