package web_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/web"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// sseHandlerTest creates a server with controllable fan-out channels.
func sseHandlerTest(t *testing.T) (*httptest.Server, *web.Server, chan plc.Event, chan plc.Snapshot) {
	t.Helper()
	rt := testRuntime(t)
	eventsCh := make(chan plc.Event, 8)
	snapsCh := make(chan plc.Snapshot, 8)

	cfg := web.Config{Enabled: true, Host: "127.0.0.1", Port: 9998}
	deps := web.Deps{
		Runtime: rt,
		CommandHandler: func(ctx context.Context, cmd plc.Command) (plc.Event, error) {
			return rt.ApplyCommand(cmd)
		},
		EventsCh:    eventsCh,
		SnapshotsCh: snapsCh,
	}
	srv := web.NewServer(cfg, deps, nil)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, srv, eventsCh, snapsCh
}

func TestSSEEndpoint_Connect(t *testing.T) {
	ts, srv, _, _ := sseHandlerTest(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	srv.PumpSSE(ctx)

	res, err := http.Get(ts.URL + "/api/stream")
	if err != nil {
		t.Fatalf("GET /api/stream: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/stream = %d, want 200", res.StatusCode)
	}
	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestSSEBroker_BroadcastReachesClient(t *testing.T) {
	ts, srv, eventsCh, _ := sseHandlerTest(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	srv.PumpSSE(ctx)

	// Connect SSE client (no timeout — body is streaming).
	req, _ := http.NewRequest("GET", ts.URL+"/api/stream", nil)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("SSE connect: %v", err)
	}
	defer resp.Body.Close()

	// Allow subscription to register before pushing.
	time.Sleep(50 * time.Millisecond)

	// Push a test event via the fan-out channel.
	eventsCh <- plc.Event{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Type:      "test_event",
		Message:   "hello from test",
	}

	// Read SSE lines until we find the event.
	lineCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				lineCh <- line
				return
			}
		}
	}()

	select {
	case line := <-lineCh:
		if !strings.Contains(line, "plc_event") {
			t.Errorf("unexpected SSE event line: %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Error("no SSE event received within 2s")
	}
}
