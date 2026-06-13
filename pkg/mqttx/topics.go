package mqttx

func StatusTopic(baseTopic string) string    { return topic(baseTopic, "status") }
func TelemetryTopic(baseTopic string) string { return topic(baseTopic, "telemetry") }
func EventsTopic(baseTopic string) string    { return topic(baseTopic, "events") }
func CommandsTopic(baseTopic string) string  { return topic(baseTopic, "commands") }
func ConfigTopic(baseTopic string) string    { return topic(baseTopic, "config") }

func topic(baseTopic, suffix string) string {
	config := Config{BaseTopic: baseTopic}.normalized()
	return config.BaseTopic + "/" + suffix
}
