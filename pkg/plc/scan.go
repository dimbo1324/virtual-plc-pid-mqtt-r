package plc

import (
	"context"
	"fmt"
	"time"
)

func (r *Runtime) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	ticker := time.NewTicker(r.config.ScanInterval)
	defer ticker.Stop()
	previousScan := time.Now()

	for {
		select {
		case <-ctx.Done():
			r.finishStopped()
			return
		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(previousScan)
			previousScan = now
			if err := r.scan(now, dt); err != nil {
				r.fault(err)
				return
			}
		}
	}
}

func (r *Runtime) scan(scanTime time.Time, dt time.Duration) error {
	started := time.Now()
	r.mu.Lock()
	for name, loop := range r.loops {
		processState := loop.process.Snapshot()
		output, err := loop.controller.Update(processState.PV, dt)
		if err != nil {
			r.mu.Unlock()
			return fmt.Errorf("update PID loop %q: %w", name, err)
		}
		if err := loop.process.ApplyMV(output); err != nil {
			r.mu.Unlock()
			return fmt.Errorf("apply output for loop %q: %w", name, err)
		}
		if _, err := loop.process.Step(dt); err != nil {
			r.mu.Unlock()
			return fmt.Errorf("step process loop %q: %w", name, err)
		}
	}

	r.snapshot.PLC.ScanCounter++
	r.snapshot = r.buildSnapshotLocked(scanTime.UTC())
	duration := time.Since(started)
	if duration <= 0 {
		duration = time.Nanosecond
	}
	r.snapshot.PLC.LastScanDurationMS = float64(duration) / float64(time.Millisecond)
	snapshot := cloneSnapshot(r.snapshot)
	r.mu.Unlock()

	r.emitSnapshot(snapshot)
	if duration >= r.config.ScanOverrunWarning {
		r.emitEvent(newEvent("warning", EventScanOverrun, "PLC scan exceeded warning threshold", map[string]any{
			"scan_duration_ms": float64(duration) / float64(time.Millisecond),
			"warning_ms":       float64(r.config.ScanOverrunWarning) / float64(time.Millisecond),
		}))
	}
	return nil
}

func (r *Runtime) buildSnapshotLocked(timestamp time.Time) Snapshot {
	loops := make(map[string]LoopSnapshot, len(r.loops))
	for name, loop := range r.loops {
		loops[name] = loop.snapshot()
	}
	return Snapshot{
		Timestamp: timestamp, DeviceID: r.config.DeviceID,
		PLC: PLCStatus{
			State: r.state, ScanIntervalMS: int(r.config.ScanInterval / time.Millisecond),
			LastScanDurationMS: r.snapshot.PLC.LastScanDurationMS, ScanCounter: r.snapshot.PLC.ScanCounter,
		},
		Loops: loops,
	}
}

func (r *Runtime) finishStopped() {
	r.mu.Lock()
	r.state = StateStopped
	r.cancel = nil
	r.done = nil
	r.snapshot = r.buildSnapshotLocked(time.Now().UTC())
	r.mu.Unlock()
	r.emitEvent(newEvent("info", EventPLCStopped, "PLC runtime stopped", nil))
}

func (r *Runtime) fault(err error) {
	r.mu.Lock()
	r.state = StateFaulted
	r.cancel = nil
	r.done = nil
	r.snapshot = r.buildSnapshotLocked(time.Now().UTC())
	r.mu.Unlock()
	r.emitEvent(newEvent("error", EventRuntimeFaulted, err.Error(), nil))
}
