package plc

import (
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/pid"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/simulator"
)

type controlLoop struct {
	name              string
	displayName       string
	unit              string
	enabled           bool
	setpointMin       float64
	setpointMax       float64
	hasSetpointLimits bool
	controller        *pid.Controller
	process           *simulator.Process
}

func newControlLoop(config LoopConfig) (*controlLoop, error) {
	pidConfig := config.PID
	pidConfig.Name = config.Name
	pidConfig.Setpoint = config.Setpoint
	pidConfig.Mode = config.Mode.pidMode()
	pidConfig.Enabled = config.Enabled
	controller, err := pid.New(pidConfig)
	if err != nil {
		return nil, err
	}

	processConfig := config.Process
	processConfig.Name = config.Name
	if processConfig.DisplayName == "" {
		processConfig.DisplayName = config.DisplayName
	}
	if processConfig.Unit == "" {
		processConfig.Unit = config.Unit
	}
	process, err := simulator.NewProcess(processConfig)
	if err != nil {
		return nil, err
	}

	displayName := config.DisplayName
	if displayName == "" {
		displayName = processConfig.DisplayName
	}
	unit := config.Unit
	if unit == "" {
		unit = processConfig.Unit
	}
	return &controlLoop{
		name: config.Name, displayName: displayName, unit: unit, enabled: config.Enabled,
		setpointMin: config.SetpointMin, setpointMax: config.SetpointMax,
		hasSetpointLimits: config.hasSetpointLimits(), controller: controller, process: process,
	}, nil
}

func (l *controlLoop) snapshot() LoopSnapshot {
	controllerState := l.controller.State()
	controllerConfig := l.controller.Config()
	processState := l.process.Snapshot()
	return LoopSnapshot{
		Name: l.name, DisplayName: l.displayName, Unit: l.unit,
		Setpoint: controllerState.Setpoint, ProcessValue: processState.PV,
		Output: controllerState.Output, Error: controllerState.Setpoint - processState.PV,
		Mode: string(controllerState.Mode), Quality: string(processState.Quality), Enabled: l.enabled,
		Kp: controllerConfig.Kp, Ki: controllerConfig.Ki, Kd: controllerConfig.Kd,
	}
}
