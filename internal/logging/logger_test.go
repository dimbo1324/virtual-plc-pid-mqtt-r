package logging_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/logging"
)

func TestNew_ReturnsLogger(t *testing.T) {
	l := logging.New("info")
	if l == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewTextLogger_WritesToOutput(t *testing.T) {
	var buf bytes.Buffer
	l := logging.NewTextLogger("info", &buf)
	if l == nil {
		t.Fatal("NewTextLogger returned nil")
	}
	l.Info("hello test")
	if !strings.Contains(buf.String(), "hello test") {
		t.Errorf("expected log output to contain %q, got: %s", "hello test", buf.String())
	}
}

func TestNewTextLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := logging.NewTextLogger("warn", &buf)
	l.Info("should be filtered")
	if strings.Contains(buf.String(), "should be filtered") {
		t.Error("info message should not appear at warn level")
	}
	l.Warn("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Errorf("warn message should appear at warn level, got: %s", buf.String())
	}
}

func TestNewTextLogger_ParseLevelCases(t *testing.T) {
	cases := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // defaults to info
		{"", slog.LevelInfo},
	}
	for _, tc := range cases {
		var buf bytes.Buffer
		l := logging.NewTextLogger(tc.input, &buf)
		// Verify effective level by checking which messages pass through.
		buf.Reset()
		l.Debug("dbg")
		l.Info("inf")
		l.Warn("wrn")
		l.Error("err")
		out := buf.String()
		switch tc.want {
		case slog.LevelDebug:
			if !strings.Contains(out, "dbg") {
				t.Errorf("level=%q: debug message missing", tc.input)
			}
		case slog.LevelInfo:
			if strings.Contains(out, "dbg") {
				t.Errorf("level=%q: debug message should be suppressed", tc.input)
			}
			if !strings.Contains(out, "inf") {
				t.Errorf("level=%q: info message missing", tc.input)
			}
		case slog.LevelWarn:
			if strings.Contains(out, "inf") {
				t.Errorf("level=%q: info message should be suppressed", tc.input)
			}
			if !strings.Contains(out, "wrn") {
				t.Errorf("level=%q: warn message missing", tc.input)
			}
		case slog.LevelError:
			if strings.Contains(out, "wrn") {
				t.Errorf("level=%q: warn message should be suppressed", tc.input)
			}
			if !strings.Contains(out, "err") {
				t.Errorf("level=%q: error message missing", tc.input)
			}
		}
	}
}
