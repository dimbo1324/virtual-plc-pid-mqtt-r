package config

import (
	"fmt"
	"strings"
)

// Validate checks the application configuration invariants.
func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("app.name must not be empty")
	}
	if strings.TrimSpace(c.App.DeviceID) == "" {
		return fmt.Errorf("app.device_id must not be empty")
	}
	if c.PLC.ScanIntervalMS <= 0 {
		return fmt.Errorf("plc.scan_interval_ms must be greater than zero")
	}
	if c.PLC.PublishIntervalMS <= 0 {
		return fmt.Errorf("plc.publish_interval_ms must be greater than zero")
	}
	if c.PLC.UIUpdateIntervalMS <= 0 {
		return fmt.Errorf("plc.ui_update_interval_ms must be greater than zero")
	}
	if c.PLC.ScanOverrunWarningMS <= 0 {
		return fmt.Errorf("plc.scan_overrun_warning_ms must be greater than zero")
	}
	if c.Web.Enabled && (c.Web.Port < 1 || c.Web.Port > 65535) {
		return fmt.Errorf("web.port must be between 1 and 65535 when web is enabled")
	}
	if c.MQTT.Enabled {
		if strings.TrimSpace(c.MQTT.BrokerURL) == "" {
			return fmt.Errorf("mqtt.broker_url must not be empty when MQTT is enabled")
		}
		if strings.TrimSpace(c.MQTT.BaseTopic) == "" {
			return fmt.Errorf("mqtt.base_topic must not be empty when MQTT is enabled")
		}
		if err := validateMQTTBaseTopic(c.MQTT.BaseTopic); err != nil {
			return err
		}
		if strings.TrimSpace(c.MQTT.ClientID) == "" {
			return fmt.Errorf("mqtt.client_id must not be empty when MQTT is enabled")
		}
		if c.MQTT.QoS > 2 {
			return fmt.Errorf("mqtt.qos must be 0, 1, or 2")
		}
		if c.MQTT.ConnectTimeoutSeconds <= 0 {
			return fmt.Errorf("mqtt.connect_timeout_seconds must be greater than zero")
		}
		if c.MQTT.ReconnectIntervalSeconds <= 0 {
			return fmt.Errorf("mqtt.reconnect_interval_seconds must be greater than zero")
		}
	}
	if c.Storage.Enabled {
		if strings.TrimSpace(c.Storage.SQLitePath) == "" {
			return fmt.Errorf("storage.sqlite_path must not be empty when storage is enabled")
		}
		if strings.TrimSpace(c.Storage.EventsJSONLPath) == "" {
			return fmt.Errorf("storage.events_jsonl_path must not be empty when storage is enabled")
		}
		if strings.TrimSpace(c.Storage.AppLogPath) == "" {
			return fmt.Errorf("storage.app_log_path must not be empty when storage is enabled")
		}
		if c.Storage.FallbackOnError {
			switch c.Storage.FallbackType {
			case "jsonl", "noop":
			default:
				return fmt.Errorf("storage.fallback_type must be \"jsonl\" or \"noop\" when fallback_on_error is true")
			}
		}
	}
	if len(c.Loops) == 0 {
		return fmt.Errorf("at least one loop must be configured")
	}

	loopNames := make(map[string]struct{}, len(c.Loops))
	for i, loop := range c.Loops {
		if err := loop.validate(i, loopNames); err != nil {
			return err
		}
	}

	return nil
}

func (l LoopConfig) validate(index int, names map[string]struct{}) error {
	name := strings.TrimSpace(l.Name)
	if name == "" {
		return fmt.Errorf("loops[%d].name must not be empty", index)
	}
	if name != l.Name {
		return fmt.Errorf("loop name %q must not have leading or trailing whitespace", l.Name)
	}
	if _, exists := names[name]; exists {
		return fmt.Errorf("loop name %q must be unique", name)
	}
	names[name] = struct{}{}

	if !validLoopMode(l.Mode) {
		return fmt.Errorf("loop %q has unsupported mode %q", name, l.Mode)
	}
	if l.SetpointMin >= l.SetpointMax {
		return fmt.Errorf("loop %q setpoint_min must be less than setpoint_max", name)
	}
	if l.Setpoint < l.SetpointMin || l.Setpoint > l.SetpointMax {
		return fmt.Errorf("loop %q setpoint must be within [%g, %g]", name, l.SetpointMin, l.SetpointMax)
	}
	if l.PID.OutputMin >= l.PID.OutputMax {
		return fmt.Errorf("loop %q pid.output_min must be less than pid.output_max", name)
	}
	if l.PID.Kp < 0 || l.PID.Ki < 0 || l.PID.Kd < 0 {
		return fmt.Errorf("loop %q PID gains must not be negative", name)
	}
	if l.Process.Min >= l.Process.Max {
		return fmt.Errorf("loop %q process.min must be less than process.max", name)
	}
	if l.Process.TauSeconds <= 0 {
		return fmt.Errorf("loop %q process.tau_seconds must be greater than zero", name)
	}
	if l.Process.InitialPV < l.Process.Min || l.Process.InitialPV > l.Process.Max {
		return fmt.Errorf("loop %q process.initial_pv must be within [%g, %g]", name, l.Process.Min, l.Process.Max)
	}
	if l.Process.NoiseStddev < 0 {
		return fmt.Errorf("loop %q process.noise_stddev must not be negative", name)
	}

	return nil
}

func validLoopMode(mode string) bool {
	switch mode {
	case "auto", "manual", "hold", "disabled":
		return true
	default:
		return false
	}
}

func validateMQTTBaseTopic(baseTopic string) error {
	topic := strings.Trim(strings.TrimSpace(baseTopic), "/")
	if strings.ContainsAny(topic, "+#\x00") {
		return fmt.Errorf("mqtt.base_topic must not contain MQTT wildcards or null bytes")
	}
	for _, level := range strings.Split(topic, "/") {
		if level == "" {
			return fmt.Errorf("mqtt.base_topic must not contain empty topic levels")
		}
	}
	return nil
}
