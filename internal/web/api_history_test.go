package web_test

import (
	"net/http"
	"testing"
)

// These tests use a nil Store (storage disabled). The handlers must return 503.

func TestRecentEvents_NoStorage(t *testing.T) {
	ts, _ := handlerTest(t) // handlerTest passes nil Store
	res, err := http.Get(ts.URL + "/api/events/recent")
	if err != nil {
		t.Fatalf("GET /api/events/recent: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 without storage, got %d", res.StatusCode)
	}
}

func TestRecentTelemetry_NoStorage(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/api/telemetry/recent?loop=test")
	if err != nil {
		t.Fatalf("GET /api/telemetry/recent: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 without storage, got %d", res.StatusCode)
	}
}

func TestRecentTelemetry_MissingLoop(t *testing.T) {
	ts, _ := handlerTest(t)
	res, err := http.Get(ts.URL + "/api/telemetry/recent")
	if err != nil {
		t.Fatalf("GET /api/telemetry/recent: %v", err)
	}
	defer res.Body.Close()
	// Storage is nil, returns 503 before the loop check.
	if res.StatusCode != http.StatusServiceUnavailable && res.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected status %d", res.StatusCode)
	}
}
