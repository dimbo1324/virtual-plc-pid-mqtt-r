package pid_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
)

func TestAutoOutputRespondsToErrorDirection(t *testing.T) {
	tests := []struct {
		name string
		pv   float64
		want func(float64) bool
	}{
		{name: "below setpoint", pv: 8, want: func(output float64) bool { return output > 0 }},
		{name: "above setpoint", pv: 12, want: func(output float64) bool { return output < 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := mustController(t, validConfig())
			output, err := controller.Update(tt.pv, time.Second)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if !tt.want(output) {
				t.Fatalf("Update() output = %v, unexpected error direction", output)
			}
		})
	}
}

func TestAutoOutputRespectsLimits(t *testing.T) {
	config := validConfig()
	config.Kp = 100
	config.Ki = 0
	config.Kd = 0

	tests := []struct {
		name string
		pv   float64
		want float64
	}{
		{name: "maximum", pv: -100, want: config.OutputMax},
		{name: "minimum", pv: 100, want: config.OutputMin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := mustController(t, config)
			output, err := controller.Update(tt.pv, time.Second)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if output != tt.want {
				t.Fatalf("Update() output = %v, want %v", output, tt.want)
			}
		})
	}
}

func TestManualModeReturnsClampedManualOutput(t *testing.T) {
	config := validConfig()
	config.Mode = pid.ModeManual
	config.OutputMin = 0
	config.OutputMax = 100
	controller := mustController(t, config)

	tests := []struct {
		name   string
		manual float64
		want   float64
	}{
		{name: "inside limits", manual: 42, want: 42},
		{name: "above maximum", manual: 150, want: 100},
		{name: "below minimum", manual: -10, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := controller.SetManualOutput(tt.manual); err != nil {
				t.Fatalf("SetManualOutput() error = %v", err)
			}
			output, err := controller.Update(5, time.Second)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if output != tt.want {
				t.Fatalf("Update() output = %v, want %v", output, tt.want)
			}
			if controller.State().Integral != 0 {
				t.Fatalf("manual mode integral = %v, want 0", controller.State().Integral)
			}
		})
	}
}

func TestSetManualOutputRejectsNonFiniteValue(t *testing.T) {
	controller := mustController(t, validConfig())

	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := controller.SetManualOutput(value); !errors.Is(err, pid.ErrNonFiniteManualOutput) {
			t.Fatalf("SetManualOutput(%v) error = %v, want %v", value, err, pid.ErrNonFiniteManualOutput)
		}
	}
}

func TestHoldModeKeepsPreviousOutput(t *testing.T) {
	controller := mustController(t, validConfig())
	previous, err := controller.Update(8, time.Second)
	if err != nil {
		t.Fatalf("initial Update() error = %v", err)
	}
	if err := controller.SetMode(pid.ModeHold); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	output, err := controller.Update(-50, time.Second)
	if err != nil {
		t.Fatalf("hold Update() error = %v", err)
	}
	if output != previous {
		t.Fatalf("hold Update() output = %v, want previous output %v", output, previous)
	}
}

func TestDisabledControllerReturnsSafeOutput(t *testing.T) {
	config := validConfig()
	config.Enabled = false
	config.Mode = pid.ModeAuto
	controller := mustController(t, config)

	output, err := controller.Update(-50, time.Second)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if output != config.OutputMin {
		t.Fatalf("Update() output = %v, want safe output %v", output, config.OutputMin)
	}
	state := controller.State()
	if state.Mode != pid.ModeDisabled || state.Enabled {
		t.Fatalf("State() = %+v, want disabled state", state)
	}
}

func TestAntiWindupStopsIntegralGrowthDuringSaturation(t *testing.T) {
	config := validConfig()
	config.Kp = 10
	config.Ki = 5
	config.Kd = 0
	config.Setpoint = 100
	config.OutputMin = 0
	config.OutputMax = 10
	controller := mustController(t, config)

	for range 100 {
		output, err := controller.Update(0, time.Second)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if output != config.OutputMax {
			t.Fatalf("Update() output = %v, want %v", output, config.OutputMax)
		}
	}

	if integral := controller.State().Integral; integral != 0 {
		t.Fatalf("Integral after sustained saturation = %v, want 0", integral)
	}
}

func TestResetClearsDynamicState(t *testing.T) {
	controller := mustController(t, validConfig())
	if _, err := controller.Update(8, time.Second); err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	if _, err := controller.Update(9, time.Second); err != nil {
		t.Fatalf("second Update() error = %v", err)
	}

	controller.Reset()
	state := controller.State()
	if state.Integral != 0 || state.Derivative != 0 || state.LastError != 0 || state.LastPV != 0 {
		t.Fatalf("State() after Reset() = %+v, dynamic values were not cleared", state)
	}
	if state.Setpoint != validConfig().Setpoint || state.Mode != pid.ModeAuto {
		t.Fatalf("State() after Reset() = %+v, configuration was not preserved", state)
	}
}

func TestUpdateRejectsInvalidInputWithoutChangingState(t *testing.T) {
	controller := mustController(t, validConfig())
	if _, err := controller.Update(8, time.Second); err != nil {
		t.Fatalf("initial Update() error = %v", err)
	}

	tests := []struct {
		name string
		pv   float64
		dt   time.Duration
		want error
	}{
		{name: "zero duration", pv: 8, dt: 0, want: pid.ErrInvalidDuration},
		{name: "negative duration", pv: 8, dt: -time.Second, want: pid.ErrInvalidDuration},
		{name: "NaN process value", pv: math.NaN(), dt: time.Second, want: pid.ErrNonFiniteProcessValue},
		{name: "infinite process value", pv: math.Inf(1), dt: time.Second, want: pid.ErrNonFiniteProcessValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := controller.State()
			output, err := controller.Update(tt.pv, tt.dt)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Update() error = %v, want %v", err, tt.want)
			}
			if output != before.Output {
				t.Fatalf("Update() output = %v, want previous output %v", output, before.Output)
			}
			if after := controller.State(); after != before {
				t.Fatalf("State() changed after invalid input: got %+v, want %+v", after, before)
			}
		})
	}
}

func TestUpdateRejectsInvalidInternalSetpoint(t *testing.T) {
	controller := mustController(t, validConfig())
	before := controller.State()
	controller.SetSetpoint(math.NaN())

	output, err := controller.Update(8, time.Second)
	if !errors.Is(err, pid.ErrNonFiniteSetpoint) {
		t.Fatalf("Update() error = %v, want %v", err, pid.ErrNonFiniteSetpoint)
	}
	if output != before.Output || controller.State() != before {
		t.Fatalf("invalid setpoint changed runtime state: got %+v, want %+v", controller.State(), before)
	}
}

func TestDerivativeOnMeasurementIsFinite(t *testing.T) {
	config := validConfig()
	config.Kp = 0
	config.Ki = 0
	config.Kd = 1
	config.Bias = 50
	config.OutputMin = 0
	config.OutputMax = 100
	controller := mustController(t, config)

	if _, err := controller.Update(10, time.Second); err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	output, err := controller.Update(11, time.Second)
	if err != nil {
		t.Fatalf("second Update() error = %v", err)
	}
	state := controller.State()
	if !almostEqual(state.Derivative, -1, 1e-12) {
		t.Fatalf("Derivative = %v, want -1", state.Derivative)
	}
	if math.IsNaN(output) || math.IsInf(output, 0) {
		t.Fatalf("Update() output = %v, want finite output", output)
	}
}

func TestStateReturnsCopy(t *testing.T) {
	controller := mustController(t, validConfig())
	if _, err := controller.Update(8, time.Second); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	copyState := controller.State()
	copyState.Integral = 999
	copyState.Output = 999

	actual := controller.State()
	if actual.Integral == 999 || actual.Output == 999 {
		t.Fatalf("State() exposed mutable internal state: %+v", actual)
	}
}

func TestSetpointUpdateAffectsFollowingOutput(t *testing.T) {
	config := validConfig()
	config.Kp = 1
	config.Ki = 0
	config.Kd = 0
	controller := mustController(t, config)

	before, err := controller.Update(10, time.Second)
	if err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	controller.SetSetpoint(20)
	after, err := controller.Update(10, time.Second)
	if err != nil {
		t.Fatalf("second Update() error = %v", err)
	}
	if after <= before {
		t.Fatalf("output after setpoint change = %v, want greater than %v", after, before)
	}
}

func TestTuningUpdateAffectsFollowingOutput(t *testing.T) {
	config := validConfig()
	config.Kp = 1
	config.Ki = 0
	config.Kd = 0
	controller := mustController(t, config)

	before, err := controller.Update(8, time.Second)
	if err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	if err := controller.SetTunings(3, 0, 0); err != nil {
		t.Fatalf("SetTunings() error = %v", err)
	}
	after, err := controller.Update(8, time.Second)
	if err != nil {
		t.Fatalf("second Update() error = %v", err)
	}
	if after <= before {
		t.Fatalf("output after tuning change = %v, want greater than %v", after, before)
	}
}

func TestModeTransitionsRemainBoundedAndFinite(t *testing.T) {
	config := validConfig()
	config.OutputMin = 0
	config.OutputMax = 100
	controller := mustController(t, config)

	if _, err := controller.Update(8, time.Second); err != nil {
		t.Fatalf("auto Update() error = %v", err)
	}
	if err := controller.SetMode(pid.ModeManual); err != nil {
		t.Fatalf("SetMode(manual) error = %v", err)
	}
	if err := controller.SetManualOutput(42); err != nil {
		t.Fatalf("SetManualOutput() error = %v", err)
	}

	transitions := []pid.Mode{pid.ModeManual, pid.ModeHold, pid.ModeDisabled, pid.ModeAuto}
	for _, mode := range transitions {
		if err := controller.SetMode(mode); err != nil {
			t.Fatalf("SetMode(%q) error = %v", mode, err)
		}
		output, err := controller.Update(8, time.Second)
		if err != nil {
			t.Fatalf("Update() in mode %q error = %v", mode, err)
		}
		if math.IsNaN(output) || math.IsInf(output, 0) || output < config.OutputMin || output > config.OutputMax {
			t.Fatalf("Update() in mode %q output = %v, want finite bounded output", mode, output)
		}
	}
}

func TestValidUpdatesNeverReturnNonFiniteOutput(t *testing.T) {
	tests := []struct {
		name string
		mode pid.Mode
		pv   float64
	}{
		{name: "auto", mode: pid.ModeAuto, pv: 8},
		{name: "manual", mode: pid.ModeManual, pv: 8},
		{name: "hold", mode: pid.ModeHold, pv: 8},
		{name: "disabled", mode: pid.ModeDisabled, pv: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig()
			config.Mode = tt.mode
			controller := mustController(t, config)
			if tt.mode == pid.ModeManual {
				if err := controller.SetManualOutput(25); err != nil {
					t.Fatalf("SetManualOutput() error = %v", err)
				}
			}
			output, err := controller.Update(tt.pv, 250*time.Millisecond)
			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if math.IsNaN(output) || math.IsInf(output, 0) {
				t.Fatalf("Update() output = %v, want finite output", output)
			}
		})
	}
}

func TestCalculationOverflowReturnsPreviousFiniteOutput(t *testing.T) {
	config := validConfig()
	config.Kp = math.MaxFloat64
	controller := mustController(t, config)
	before := controller.State()

	output, err := controller.Update(-math.MaxFloat64, time.Second)
	if !errors.Is(err, pid.ErrNonFiniteCalculation) {
		t.Fatalf("Update() error = %v, want %v", err, pid.ErrNonFiniteCalculation)
	}
	if output != before.Output || math.IsNaN(output) || math.IsInf(output, 0) {
		t.Fatalf("Update() output = %v, want previous finite output %v", output, before.Output)
	}
	if controller.State() != before {
		t.Fatalf("State() changed after overflow: got %+v, want %+v", controller.State(), before)
	}
}
