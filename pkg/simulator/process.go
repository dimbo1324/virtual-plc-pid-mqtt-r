package simulator

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Process models one synthetic first-order process variable.
type Process struct {
	config      Config
	pv          float64
	mv          float64
	target      float64
	quality     Quality
	disturbance *Disturbance
	random      *rand.Rand
}

// NewProcess validates config and creates an independent process model.
func NewProcess(config Config) (*Process, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	config = config.normalized()

	process := &Process{
		config:  config,
		pv:      config.InitialPV,
		quality: QualityGood,
		random:  newRandomSource(config.RandomSeed),
	}
	target, err := process.targetFor(0, nil)
	if err != nil {
		return nil, fmt.Errorf("initialize process target: %w", err)
	}
	process.target = target

	return process, nil
}

// ApplyMV sets the manipulated value used by subsequent simulation steps.
func (p *Process) ApplyMV(mv float64) error {
	if !finite(mv) {
		return ErrInvalidMV
	}
	target, err := p.targetFor(mv, p.disturbance)
	if err != nil {
		return err
	}
	p.mv = mv
	p.target = target
	return nil
}

// Step advances the process using first-order dynamics and optional noise.
func (p *Process) Step(dt time.Duration) (Snapshot, error) {
	if dt <= 0 {
		return p.Snapshot(), ErrInvalidDuration
	}
	if err := p.validateState(); err != nil {
		return p.Snapshot(), err
	}

	seconds := dt.Seconds()
	target, err := p.targetFor(p.mv, p.disturbance)
	if err != nil {
		return p.Snapshot(), err
	}
	nextPV := firstOrderStep(p.pv, target, seconds, p.config.TauSeconds)
	if !finite(nextPV) {
		return p.Snapshot(), ErrNonFiniteCalculation
	}

	noise := sampleNoise(p.random, p.config.NoiseStddev)
	if !finite(noise) || !finite(nextPV+noise) {
		return p.Snapshot(), ErrNonFiniteCalculation
	}
	nextPV += noise
	nextPV, clamped := clamp(nextPV, p.config.Min, p.config.Max)

	nextDisturbance := advanceDisturbance(p.disturbance, seconds)
	nextTarget, err := p.targetFor(p.mv, nextDisturbance)
	if err != nil {
		return p.Snapshot(), err
	}

	p.pv = nextPV
	p.target = nextTarget
	p.disturbance = nextDisturbance
	p.quality = QualityGood
	if clamped {
		p.quality = QualityUncertain
	}

	return p.Snapshot(), nil
}

// Snapshot returns a copy of the current process state.
func (p *Process) Snapshot() Snapshot {
	return Snapshot{
		Name:              p.config.Name,
		DisplayName:       p.config.DisplayName,
		Unit:              p.config.Unit,
		PV:                p.pv,
		MV:                p.mv,
		Target:            p.target,
		Min:               p.config.Min,
		Max:               p.config.Max,
		Quality:           p.quality,
		DisturbanceActive: p.disturbance != nil,
	}
}

// InjectDisturbance replaces the active disturbance with d.
func (p *Process) InjectDisturbance(d Disturbance) error {
	if err := d.validate(); err != nil {
		return err
	}
	d.Name = strings.TrimSpace(d.Name)
	target, err := p.targetFor(p.mv, &d)
	if err != nil {
		return err
	}
	p.disturbance = &d
	p.target = target
	return nil
}

// ClearDisturbance removes the active disturbance immediately.
func (p *Process) ClearDisturbance() {
	p.disturbance = nil
	target, err := p.targetFor(p.mv, nil)
	if err == nil {
		p.target = target
	}
}

// Reset restores the initial process value and deterministic random sequence.
func (p *Process) Reset() {
	p.pv = p.config.InitialPV
	p.mv = 0
	p.disturbance = nil
	p.quality = QualityGood
	p.random = newRandomSource(p.config.RandomSeed)
	p.target = p.config.Base
}

// Config returns a copy of the process configuration.
func (p *Process) Config() Config {
	return p.config
}

func (p *Process) targetFor(mv float64, disturbance *Disturbance) (float64, error) {
	target := p.config.Base + p.config.Gain*mv
	if !finite(target) {
		return 0, ErrNonFiniteCalculation
	}
	if disturbance != nil {
		target += disturbance.Amplitude
		if !finite(target) {
			return 0, ErrNonFiniteCalculation
		}
	}
	return target, nil
}

func (p *Process) validateState() error {
	if !finite(p.pv) || !finite(p.mv) || !finite(p.target) || !p.quality.Valid() {
		return ErrInvalidState
	}
	if p.disturbance != nil {
		if err := p.disturbance.validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidState, err)
		}
	}
	return nil
}

func advanceDisturbance(current *Disturbance, seconds float64) *Disturbance {
	if current == nil {
		return nil
	}
	next := *current
	next.RemainingSeconds -= seconds
	if next.RemainingSeconds <= 0 {
		return nil
	}
	return &next
}
