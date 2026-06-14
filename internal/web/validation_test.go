package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryInt(t *testing.T) {
	cases := []struct {
		query         string
		key           string
		def, min, max int
		want          int
	}{
		{"", "limit", 10, 1, 100, 10},
		{"limit=50", "limit", 10, 1, 100, 50},
		{"limit=0", "limit", 10, 1, 100, 1},
		{"limit=200", "limit", 10, 1, 100, 100},
		{"limit=abc", "limit", 10, 1, 100, 10},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("GET", "/?"+tc.query, nil)
		got := queryInt(req, tc.key, tc.def, tc.min, tc.max)
		if got != tc.want {
			t.Errorf("queryInt(q=%q, key=%q, def=%d, min=%d, max=%d) = %d, want %d",
				tc.query, tc.key, tc.def, tc.min, tc.max, got, tc.want)
		}
	}
}

func TestDecodeJSON_Valid(t *testing.T) {
	body := bytes.NewBufferString(`{"loop":"pressure","setpoint":6.0}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/json")

	var v map[string]any
	if err := decodeJSON(req, &v); err != nil {
		t.Fatalf("decodeJSON: %v", err)
	}
	if v["loop"] != "pressure" {
		t.Errorf("unexpected loop: %v", v["loop"])
	}
}

func TestDecodeJSON_Invalid(t *testing.T) {
	body := bytes.NewBufferString("not json")
	req := httptest.NewRequest("POST", "/", body)
	var v map[string]any
	if err := decodeJSON(req, &v); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDecodeJSON_UnknownField(t *testing.T) {
	body := bytes.NewBufferString(`{"unknown_field":"value"}`)
	req := httptest.NewRequest("POST", "/", body)
	var v struct{ Known string }
	if err := decodeJSON(req, &v); err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestJSONOK(t *testing.T) {
	w := httptest.NewRecorder()
	jsonOK(w, map[string]string{"status": "ok"})
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	jsonError(w, http.StatusBadRequest, "bad input")
	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", res.StatusCode)
	}
}
