package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Config describes the future runtime while keeping Stage 01 limited to
// loading and validation.
type Config struct {
	App     AppConfig     `json:"app"`
	PLC     PLCConfig     `json:"plc"`
	MQTT    MQTTConfig    `json:"mqtt"`
	Web     WebConfig     `json:"web"`
	Storage StorageConfig `json:"storage"`
	Loops   []LoopConfig  `json:"loops"`
}

type AppConfig struct {
	Name        string `json:"name"`
	DeviceID    string `json:"device_id"`
	AutoStart   bool   `json:"auto_start"`
	OpenBrowser bool   `json:"open_browser"`
}

type PLCConfig struct {
	ScanIntervalMS       int `json:"scan_interval_ms"`
	PublishIntervalMS    int `json:"publish_interval_ms"`
	UIUpdateIntervalMS   int `json:"ui_update_interval_ms"`
	ScanOverrunWarningMS int `json:"scan_overrun_warning_ms"`
}

type MQTTConfig struct {
	Enabled                  bool   `json:"enabled"`
	BrokerURL                string `json:"broker_url"`
	ClientID                 string `json:"client_id"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	BaseTopic                string `json:"base_topic"`
	QoS                      byte   `json:"qos"`
	ConnectTimeoutSeconds    int    `json:"connect_timeout_seconds"`
	ReconnectIntervalSeconds int    `json:"reconnect_interval_seconds"`
}

type WebConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

type StorageConfig struct {
	Enabled             bool   `json:"enabled"`
	Type                string `json:"type"`
	SQLitePath          string `json:"sqlite_path"`
	EventsJSONLPath     string `json:"events_jsonl_path"`
	AppLogPath          string `json:"app_log_path"`
	RetentionMaxSamples int    `json:"retention_max_samples"`
}

type LoopConfig struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Unit        string        `json:"unit"`
	Enabled     bool          `json:"enabled"`
	Mode        string        `json:"mode"`
	Setpoint    float64       `json:"setpoint"`
	SetpointMin float64       `json:"setpoint_min"`
	SetpointMax float64       `json:"setpoint_max"`
	PID         PIDConfig     `json:"pid"`
	Process     ProcessConfig `json:"process"`
}

type PIDConfig struct {
	Kp        float64 `json:"kp"`
	Ki        float64 `json:"ki"`
	Kd        float64 `json:"kd"`
	Bias      float64 `json:"bias"`
	OutputMin float64 `json:"output_min"`
	OutputMax float64 `json:"output_max"`
}

type ProcessConfig struct {
	InitialPV          float64 `json:"initial_pv"`
	Min                float64 `json:"min"`
	Max                float64 `json:"max"`
	Base               float64 `json:"base"`
	Gain               float64 `json:"gain"`
	TauSeconds         float64 `json:"tau_seconds"`
	NoiseStddev        float64 `json:"noise_stddev"`
	RandomDisturbances bool    `json:"random_disturbances"`
}

// Load reads a JSON configuration file from path.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

	if err := ensureEndOfJSON(decoder); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

	return cfg, nil
}

func ensureEndOfJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
