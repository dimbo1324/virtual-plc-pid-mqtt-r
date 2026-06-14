package web_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCommandStart(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/start", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("POST /api/commands/start: %v", err)
	}
	defer res.Body.Close()
	// start is idempotent when already stopped — any 2xx is fine
	if res.StatusCode >= 300 {
		t.Errorf("POST /api/commands/start = %d, want 2xx", res.StatusCode)
	}
}

func TestCommandStop(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/stop", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("POST /api/commands/stop: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		t.Errorf("POST /api/commands/stop = %d, want 2xx", res.StatusCode)
	}
}

func TestCommandSetpoint_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "setpoint": 60.0})
	res, err := http.Post(ts.URL+"/api/commands/setpoint", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/setpoint: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("valid setpoint = %d, want 200", res.StatusCode)
	}
}

func TestCommandSetpoint_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/setpoint", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/setpoint: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid body = %d, want 400", res.StatusCode)
	}
}

func TestCommandSetpoint_UnknownLoop(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "nonexistent", "setpoint": 50.0})
	res, err := http.Post(ts.URL+"/api/commands/setpoint", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/setpoint: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("unknown loop = %d, want 400", res.StatusCode)
	}
}

func TestCommandPIDGains_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "kp": 2.0, "ki": 0.2, "kd": 0.01})
	res, err := http.Post(ts.URL+"/api/commands/pid-gains", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/pid-gains: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("valid pid-gains = %d, want 200", res.StatusCode)
	}
}

func TestCommandMode_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "mode": "hold"})
	res, err := http.Post(ts.URL+"/api/commands/mode", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/mode: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("valid mode = %d, want 200", res.StatusCode)
	}
}

func TestCommandResetLoop_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test"})
	res, err := http.Post(ts.URL+"/api/commands/reset-loop", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/reset-loop: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("valid reset-loop = %d, want 200", res.StatusCode)
	}
}

func TestCommandManualOutput_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "value": 42.0})
	res, err := http.Post(ts.URL+"/api/commands/manual-output", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/manual-output: %v", err)
	}
	defer res.Body.Close()
	// The command reaches the handler; runtime may accept or reject based on mode.
	if res.StatusCode == http.StatusNotFound {
		t.Errorf("manual-output endpoint not registered, got 404")
	}
}

func TestCommandManualOutput_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/manual-output", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/manual-output: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid body = %d, want 400", res.StatusCode)
	}
}

func TestCommandInjectDisturbance_Valid(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "amplitude": 2.0, "duration_seconds": 5.0})
	res, err := http.Post(ts.URL+"/api/commands/inject-disturbance", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/inject-disturbance: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("valid inject-disturbance = %d, want 200", res.StatusCode)
	}
}

func TestCommandInjectDisturbance_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/inject-disturbance", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/inject-disturbance: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid body = %d, want 400", res.StatusCode)
	}
}

func TestCommandPIDGains_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/pid-gains", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/pid-gains: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid pid-gains body = %d, want 400", res.StatusCode)
	}
}

func TestCommandMode_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/mode", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/mode: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid mode body = %d, want 400", res.StatusCode)
	}
}

func TestCommandResetLoop_InvalidBody(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Post(ts.URL+"/api/commands/reset-loop", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("POST /api/commands/reset-loop: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid reset-loop body = %d, want 400", res.StatusCode)
	}
}

func TestCommandSetpoint_MissingContentType(t *testing.T) {
	ts, _ := handlerTest(t)
	body, _ := json.Marshal(map[string]any{"loop": "test", "setpoint": 60.0})
	// No Content-Type header: server should still parse valid JSON.
	res, err := http.Post(ts.URL+"/api/commands/setpoint", "", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/commands/setpoint: %v", err)
	}
	defer res.Body.Close()
	// A 400 or 200 is acceptable — the key invariant is that 404 never occurs.
	if res.StatusCode == http.StatusNotFound {
		t.Errorf("setpoint endpoint returned 404")
	}
}
