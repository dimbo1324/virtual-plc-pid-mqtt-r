package plc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	defaultEventBuffer    = 64
	defaultSnapshotBuffer = 16
)

// Runtime owns control loops and executes their PLC scan cycle.
type Runtime struct {
	mu sync.RWMutex

	config      Config
	loops       map[string]*controlLoop
	externalPVs map[string]float64 // set by InjectPV; consumed each scan cycle
	state       State
	snapshot    Snapshot

	cancel context.CancelFunc
	done   chan struct{}

	events    chan Event
	snapshots chan Snapshot
}

// NewRuntime validates config and creates a stopped runtime.
func NewRuntime(config Config) (*Runtime, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	loops := make(map[string]*controlLoop, len(config.Loops))
	for _, loopConfig := range config.Loops {
		loop, err := newControlLoop(loopConfig)
		if err != nil {
			return nil, fmt.Errorf("create loop %q: %w", loopConfig.Name, err)
		}
		loops[loop.name] = loop
	}

	runtime := &Runtime{
		config: config, loops: loops, state: StateStopped,
		externalPVs: make(map[string]float64),
		events:      make(chan Event, defaultEventBuffer),
		snapshots:   make(chan Snapshot, defaultSnapshotBuffer),
	}
	runtime.snapshot = runtime.buildSnapshotLocked(time.Now().UTC())
	return runtime, nil
}

// Start launches the scan goroutine and returns immediately.
func (r *Runtime) Start(parent context.Context) error {
	if parent == nil {
		parent = context.Background()
	}
	if err := parent.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	switch r.state {
	case StateRunning, StateStarting:
		r.mu.Unlock()
		return nil
	case StateStopping:
		r.mu.Unlock()
		return fmt.Errorf("%w: runtime is stopping", ErrInvalidState)
	}
	r.state = StateStarting
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})
	r.cancel = cancel
	r.done = done
	r.state = StateRunning
	r.snapshot = r.buildSnapshotLocked(time.Now().UTC())
	r.mu.Unlock()

	r.emitEvent(newEvent("info", EventPLCStarted, "PLC runtime started", nil))
	go r.run(ctx, done)
	return nil
}

// Stop cancels the scan goroutine and waits for it to finish.
func (r *Runtime) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	if r.state == StateStopped {
		r.mu.Unlock()
		return nil
	}
	if r.state == StateFaulted && r.done == nil {
		r.mu.Unlock()
		return nil
	}
	done := r.done
	if r.state != StateStopping {
		r.state = StateStopping
		if r.cancel != nil {
			r.cancel()
		}
		r.snapshot = r.buildSnapshotLocked(time.Now().UTC())
	}
	r.mu.Unlock()

	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// State returns the current runtime lifecycle state.
func (r *Runtime) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// Snapshot returns a deep copy of the latest runtime snapshot.
func (r *Runtime) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneSnapshot(r.snapshot)
}

// Events returns the buffered runtime event stream.
func (r *Runtime) Events() <-chan Event { return r.events }

// Snapshots returns the buffered scan snapshot stream.
func (r *Runtime) Snapshots() <-chan Snapshot { return r.snapshots }

// InjectPV supplies an externally sourced process variable for a named loop.
// On the next scan tick this value is used in place of the simulator output for
// the PID controller update. Returns false if the loop does not exist.
// This is the hook for pkg/input.Provider adapters (OPC-UA, Modbus, REST, etc.).
func (r *Runtime) InjectPV(name string, pv float64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loops[name]; !ok {
		return false
	}
	r.externalPVs[name] = pv
	return true
}

func (r *Runtime) emitEvent(event Event) {
	select {
	case r.events <- event:
	default:
	}
}

func (r *Runtime) emitSnapshot(snapshot Snapshot) {
	select {
	case r.snapshots <- cloneSnapshot(snapshot):
	default:
	}
}
