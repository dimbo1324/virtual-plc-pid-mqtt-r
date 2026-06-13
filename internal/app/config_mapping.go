package app

import (
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/mqttx"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

func mapPLCConfig(config config.Config) plc.Config {
	loops := make([]plc.LoopConfig, 0, len(config.Loops))
	for index, loop := range config.Loops {
		loops = append(loops, plc.LoopConfig{
			Name: loop.Name, DisplayName: loop.DisplayName, Unit: loop.Unit,
			Enabled: loop.Enabled, Mode: plc.LoopMode(loop.Mode), Setpoint: loop.Setpoint,
			SetpointMin: loop.SetpointMin, SetpointMax: loop.SetpointMax,
			PID: pid.Config{
				Name: loop.Name, Kp: loop.PID.Kp, Ki: loop.PID.Ki, Kd: loop.PID.Kd,
				Bias: loop.PID.Bias, OutputMin: loop.PID.OutputMin, OutputMax: loop.PID.OutputMax,
				Setpoint: loop.Setpoint, Mode: pid.Mode(loop.Mode), Enabled: loop.Enabled,
			},
			Process: simulator.Config{
				Name: loop.Name, DisplayName: loop.DisplayName, Unit: loop.Unit,
				InitialPV: loop.Process.InitialPV, Min: loop.Process.Min, Max: loop.Process.Max,
				Base: loop.Process.Base, Gain: loop.Process.Gain, TauSeconds: loop.Process.TauSeconds,
				NoiseStddev: loop.Process.NoiseStddev, RandomSeed: int64(index + 1),
				RandomDisturbances: loop.Process.RandomDisturbances,
			},
		})
	}
	return plc.Config{
		DeviceID:           config.App.DeviceID,
		ScanInterval:       time.Duration(config.PLC.ScanIntervalMS) * time.Millisecond,
		PublishInterval:    time.Duration(config.PLC.PublishIntervalMS) * time.Millisecond,
		UIUpdateInterval:   time.Duration(config.PLC.UIUpdateIntervalMS) * time.Millisecond,
		ScanOverrunWarning: time.Duration(config.PLC.ScanOverrunWarningMS) * time.Millisecond,
		Loops:              loops,
	}
}

func mapMQTTConfig(config config.Config) mqttx.Config {
	return mqttx.Config{
		Enabled: config.MQTT.Enabled, BrokerURL: config.MQTT.BrokerURL,
		ClientID: config.MQTT.ClientID, Username: config.MQTT.Username, Password: config.MQTT.Password,
		BaseTopic: config.MQTT.BaseTopic, QoS: config.MQTT.QoS,
		ConnectTimeout:    time.Duration(config.MQTT.ConnectTimeoutSeconds) * time.Second,
		ReconnectInterval: time.Duration(config.MQTT.ReconnectIntervalSeconds) * time.Second,
	}
}
