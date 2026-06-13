package config

// Default returns the Stage 01 baseline configuration.
func Default() Config {
	return Config{
		App: AppConfig{
			Name:        "virtual-plc-pid-mqtt-r",
			DeviceID:    "vplc_001",
			AutoStart:   true,
			OpenBrowser: false,
		},
		PLC: PLCConfig{
			ScanIntervalMS:       500,
			PublishIntervalMS:    1000,
			UIUpdateIntervalMS:   250,
			ScanOverrunWarningMS: 500,
		},
		MQTT: MQTTConfig{
			Enabled:                  true,
			BrokerURL:                "tcp://localhost:1883",
			ClientID:                 "virtual-plc-pid-mqtt-r-vplc-001",
			BaseTopic:                "vplc/vplc_001",
			QoS:                      0,
			ConnectTimeoutSeconds:    5,
			ReconnectIntervalSeconds: 3,
		},
		Web: WebConfig{
			Enabled: true,
			Host:    "127.0.0.1",
			Port:    8080,
		},
		Storage: StorageConfig{
			Enabled:             true,
			Type:                "sqlite",
			SQLitePath:          "data/history.db",
			EventsJSONLPath:     "logs/events.jsonl",
			AppLogPath:          "logs/app.log",
			RetentionMaxSamples: 100000,
		},
		Loops: []LoopConfig{
			{
				Name: "pressure", DisplayName: "Pressure", Unit: "bar", Enabled: true,
				Mode: "auto", Setpoint: 6, SetpointMin: 0, SetpointMax: 12,
				PID: PIDConfig{Kp: 3, Ki: 0.25, Kd: 0.05, OutputMin: 0, OutputMax: 100},
				Process: ProcessConfig{
					InitialPV: 4, Min: 0, Max: 12, Gain: 0.10, TauSeconds: 15,
					NoiseStddev: 0.03, RandomDisturbances: true,
				},
			},
			{
				Name: "temperature", DisplayName: "Temperature", Unit: "C", Enabled: true,
				Mode: "auto", Setpoint: 180, SetpointMin: 0, SetpointMax: 250,
				PID: PIDConfig{Kp: 1.8, Ki: 0.10, Kd: 0.02, OutputMin: 0, OutputMax: 100},
				Process: ProcessConfig{
					InitialPV: 80, Min: 0, Max: 250, Base: 20, Gain: 2, TauSeconds: 60,
					NoiseStddev: 0.2, RandomDisturbances: true,
				},
			},
			{
				Name: "level", DisplayName: "Level", Unit: "%", Enabled: true,
				Mode: "auto", Setpoint: 50, SetpointMin: 0, SetpointMax: 100,
				PID: PIDConfig{Kp: 2.5, Ki: 0.15, Kd: 0.03, OutputMin: 0, OutputMax: 100},
				Process: ProcessConfig{
					InitialPV: 45, Min: 0, Max: 100, Gain: 1, TauSeconds: 25,
					NoiseStddev: 0.1, RandomDisturbances: true,
				},
			},
		},
	}
}
