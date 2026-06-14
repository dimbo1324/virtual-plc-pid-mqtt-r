package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/web"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func testRuntime(t *testing.T) *plc.Runtime {
	t.Helper()
	cfg := plc.Config{
		DeviceID:           "test_device",
		ScanInterval:       100 * time.Millisecond,
		PublishInterval:    500 * time.Millisecond,
		UIUpdateInterval:   200 * time.Millisecond,
		ScanOverrunWarning: 200 * time.Millisecond,
		Loops: []plc.LoopConfig{{
			Name: "test", DisplayName: "Test", Unit: "bar",
			Enabled: true, Mode: plc.LoopModeAuto, Setpoint: 50,
			SetpointMin: 0, SetpointMax: 100,
			PID: pid.Config{
				Name: "test", Kp: 1, Ki: 0.1, Kd: 0,
				OutputMin: 0, OutputMax: 100, Setpoint: 50, Mode: pid.ModeAuto, Enabled: true,
			},
			Process: simulator.Config{
				Name: "test", Min: 0, Max: 100,
				Gain: 1, TauSeconds: 10, InitialPV: 45,
			},
		}},
	}
	rt, err := plc.NewRuntime(cfg)
	if err != nil {
		t.Fatalf("testRuntime: %v", err)
	}
	return rt
}

func testServer(t *testing.T) *web.Server {
	t.Helper()
	rt := testRuntime(t)
	cfg := web.Config{Enabled: true, Host: "127.0.0.1", Port: 8181}
	deps := web.Deps{
		Runtime: rt,
		CommandHandler: func(ctx context.Context, cmd plc.Command) (plc.Event, error) {
			return rt.ApplyCommand(cmd)
		},
		EventsCh:    make(chan plc.Event, 8),
		SnapshotsCh: make(chan plc.Snapshot, 8),
	}
	return web.NewServer(cfg, deps, nil)
}

func TestNewServer_Addr(t *testing.T) {
	s := testServer(t)
	if s.Addr() != "127.0.0.1:8181" {
		t.Errorf("unexpected addr: %s", s.Addr())
	}
}

func TestConfig_Validate(t *testing.T) {
	cases := []struct {
		cfg     web.Config
		wantErr bool
	}{
		{web.Config{Enabled: false}, false},
		{web.Config{Enabled: true, Host: "127.0.0.1", Port: 8080}, false},
		{web.Config{Enabled: true, Host: "", Port: 8080}, true},
		{web.Config{Enabled: true, Host: "127.0.0.1", Port: 0}, true},
		{web.Config{Enabled: true, Host: "127.0.0.1", Port: 99999}, true},
	}
	for _, tc := range cases {
		err := tc.cfg.Validate()
		if (err != nil) != tc.wantErr {
			t.Errorf("Validate(%+v) error=%v wantErr=%v", tc.cfg, err, tc.wantErr)
		}
	}
}

// handlerTest wires a Server's routes into an httptest.Server without binding
// a real TCP port.
func handlerTest(t *testing.T) (*httptest.Server, *plc.Runtime) {
	t.Helper()
	rt := testRuntime(t)
	cfg := web.Config{Enabled: true, Host: "127.0.0.1", Port: 9999}
	deps := web.Deps{
		Runtime: rt,
		CommandHandler: func(ctx context.Context, cmd plc.Command) (plc.Event, error) {
			return rt.ApplyCommand(cmd)
		},
		EventsCh:    make(chan plc.Event, 8),
		SnapshotsCh: make(chan plc.Snapshot, 8),
	}
	srv := web.NewServer(cfg, deps, nil)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, rt
}

func TestServeStaticIndex(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET / = %d, want 200", res.StatusCode)
	}
}

func TestGetStatus(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("GET /api/status: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /api/status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestGetSnapshot(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/api/snapshot")
	if err != nil {
		t.Fatalf("GET /api/snapshot: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /api/snapshot = %d, want 200", res.StatusCode)
	}
}

func TestServer_StartShutdown(t *testing.T) {
	rt := testRuntime(t)
	// Use a distinct port to avoid colliding with testServer's 8181.
	cfg := web.Config{Enabled: true, Host: "127.0.0.1", Port: 8182}
	deps := web.Deps{
		Runtime: rt,
		CommandHandler: func(ctx context.Context, cmd plc.Command) (plc.Event, error) {
			return rt.ApplyCommand(cmd)
		},
		EventsCh:    make(chan plc.Event, 4),
		SnapshotsCh: make(chan plc.Snapshot, 4),
	}
	s := web.NewServer(cfg, deps, nil)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- s.Start(ctx) }()

	// Allow the server to bind before cancelling.
	time.Sleep(40 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		// Start returns nil when ctx is cancelled (graceful shutdown path).
		if err != nil {
			t.Errorf("Start() = %v, want nil after cancel", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not return after context cancel")
	}
}

func TestServer_Addr(t *testing.T) {
	s := testServer(t)
	addr := s.Addr()
	if addr == "" {
		t.Error("Addr() returned empty string")
	}
}

func TestHealthz(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /healthz = %d, want 200", res.StatusCode)
	}
}

func TestReadyz_NotRunning(t *testing.T) {
	ts, _ := handlerTest(t)
	// PLC is not started in tests, so /readyz should return 503.
	res, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("GET /readyz = %d, want 503 (PLC not started)", res.StatusCode)
	}
}

func TestReadyz_Running(t *testing.T) {
	ts, rt := handlerTest(t)
	if err := rt.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = rt.Stop(context.Background()) })

	res, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /readyz = %d, want 200 (PLC running)", res.StatusCode)
	}
}
