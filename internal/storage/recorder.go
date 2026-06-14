package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// drainTimeout is the maximum time Stop waits for the queue to drain.
const drainTimeout = 3 * time.Second

// retentionInterval controls how often retention cleanup runs in the worker.
const retentionInterval = 5 * time.Minute

type recorderMsg struct {
	kind    msgKind
	snap    plc.Snapshot
	event   EventRecord
	command CommandRecord
	pid     PIDChangeRecord
}

type msgKind int

const (
	kindSnapshot msgKind = iota
	kindEvent
	kindCommand
	kindPIDChange
)

// Recorder is a bounded async writer that bridges PLC runtime channels to the
// storage layer without blocking the scan loop. Submissions that would overflow
// the queue are silently dropped (non-blocking policy).
type Recorder struct {
	store  *Store
	jsonl  *JSONLWriter
	queue  chan recorderMsg
	logger *slog.Logger
}

// NewRecorder creates a Recorder. queueSize controls the bounded channel depth.
// logger may be nil; a default no-op logger is used in that case.
func NewRecorder(store *Store, jsonl *JSONLWriter, queueSize int, logger *slog.Logger) *Recorder {
	if logger == nil {
		logger = slog.Default()
	}
	return &Recorder{
		store:  store,
		jsonl:  jsonl,
		queue:  make(chan recorderMsg, queueSize),
		logger: logger,
	}
}

// Start launches the worker goroutine. It returns when ctx is cancelled.
func (r *Recorder) Start(ctx context.Context) {
	go r.worker(ctx)
}

// Stop drains the queue for up to drainTimeout, then returns.
func (r *Recorder) Stop(_ context.Context) error {
	deadline := time.Now().Add(drainTimeout)
	for {
		if len(r.queue) == 0 || time.Now().After(deadline) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// RecordSnapshot enqueues a snapshot for telemetry persistence. Returns false
// if the queue is full (non-blocking; the scan loop must never be held).
func (r *Recorder) RecordSnapshot(snap plc.Snapshot) bool {
	// Clone the loops map so the scan loop can mutate the original safely.
	cloned := plc.Snapshot{
		Timestamp: snap.Timestamp,
		DeviceID:  snap.DeviceID,
		PLC:       snap.PLC,
		Loops:     make(map[string]plc.LoopSnapshot, len(snap.Loops)),
	}
	for k, v := range snap.Loops {
		cloned.Loops[k] = v
	}
	return r.enqueue(recorderMsg{kind: kindSnapshot, snap: cloned})
}

// RecordEvent enqueues an event for SQLite and JSONL persistence.
func (r *Recorder) RecordEvent(event EventRecord) bool {
	return r.enqueue(recorderMsg{kind: kindEvent, event: event})
}

// RecordCommand enqueues a command record for persistence.
func (r *Recorder) RecordCommand(record CommandRecord) bool {
	return r.enqueue(recorderMsg{kind: kindCommand, command: record})
}

// RecordPIDChange enqueues a PID tuning change for persistence.
func (r *Recorder) RecordPIDChange(record PIDChangeRecord) bool {
	return r.enqueue(recorderMsg{kind: kindPIDChange, pid: record})
}

func (r *Recorder) enqueue(msg recorderMsg) bool {
	select {
	case r.queue <- msg:
		return true
	default:
		return false
	}
}

func (r *Recorder) worker(ctx context.Context) {
	retentionTicker := time.NewTicker(retentionInterval)
	defer retentionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Drain remaining items within drainTimeout.
			drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
			defer cancel()
			r.drainQueue(drainCtx)
			return

		case <-retentionTicker.C:
			if err := r.store.ApplyRetention(ctx); err != nil {
				r.logger.Warn("storage retention error", "error", err)
			}

		case msg := <-r.queue:
			r.processMsg(ctx, msg)
		}
	}
}

func (r *Recorder) drainQueue(ctx context.Context) {
	for {
		select {
		case msg := <-r.queue:
			r.processMsg(ctx, msg)
		case <-ctx.Done():
			return
		default:
			return
		}
	}
}

func (r *Recorder) processMsg(ctx context.Context, msg recorderMsg) {
	switch msg.kind {
	case kindSnapshot:
		r.persistSnapshot(ctx, msg.snap)
	case kindEvent:
		r.persistEvent(ctx, msg.event)
	case kindCommand:
		if err := r.store.InsertCommand(ctx, msg.command); err != nil {
			r.logger.Warn("storage insert command", "error", err)
		}
	case kindPIDChange:
		if err := r.store.InsertPIDChange(ctx, msg.pid); err != nil {
			r.logger.Warn("storage insert pid change", "error", err)
		}
	}
}

func (r *Recorder) persistSnapshot(ctx context.Context, snap plc.Snapshot) {
	samples := make([]TelemetrySample, 0, len(snap.Loops))
	for _, loop := range snap.Loops {
		samples = append(samples, TelemetrySample{
			Timestamp:   snap.Timestamp,
			DeviceID:    snap.DeviceID,
			ScanCounter: snap.PLC.ScanCounter,
			LoopName:    loop.Name,
			SP:          loop.Setpoint,
			PV:          loop.ProcessValue,
			MV:          loop.Output,
			Error:       loop.Error,
			Mode:        loop.Mode,
			Quality:     loop.Quality,
			Unit:        loop.Unit,
		})
	}
	if err := r.store.InsertTelemetrySamples(ctx, samples); err != nil {
		r.logger.Warn("storage insert telemetry", "error", err)
	}
}

func (r *Recorder) persistEvent(ctx context.Context, event EventRecord) {
	if err := r.store.InsertEvent(ctx, event); err != nil {
		r.logger.Warn("storage insert event", "error", err)
	}
	if r.jsonl != nil {
		if err := r.jsonl.WriteEvent(ctx, event); err != nil {
			r.logger.Warn("jsonl write event", "error", err)
		}
	}
}
