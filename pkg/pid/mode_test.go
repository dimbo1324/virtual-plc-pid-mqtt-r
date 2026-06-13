package pid_test

import (
	"errors"
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
)

func TestModeValid(t *testing.T) {
	tests := []struct {
		mode pid.Mode
		want bool
	}{
		{mode: pid.ModeAuto, want: true},
		{mode: pid.ModeManual, want: true},
		{mode: pid.ModeHold, want: true},
		{mode: pid.ModeDisabled, want: true},
		{mode: pid.Mode("cascade"), want: false},
		{mode: pid.Mode(""), want: false},
	}

	for _, tt := range tests {
		if got := tt.mode.Valid(); got != tt.want {
			t.Errorf("Mode(%q).Valid() = %v, want %v", tt.mode, got, tt.want)
		}
	}
}

func TestSetModeRejectsUnknownModeWithoutChangingState(t *testing.T) {
	controller := mustController(t, validConfig())
	before := controller.State()

	err := controller.SetMode(pid.Mode("cascade"))
	if !errors.Is(err, pid.ErrInvalidMode) {
		t.Fatalf("SetMode() error = %v, want %v", err, pid.ErrInvalidMode)
	}
	if got := controller.State(); got != before {
		t.Fatalf("State() changed after invalid mode: got %+v, want %+v", got, before)
	}
}
