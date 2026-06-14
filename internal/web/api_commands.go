package web

import (
	"net/http"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// requireLoop returns false and writes a 400 JSON error if loop is empty.
func requireLoop(w http.ResponseWriter, loop string) bool {
	if loop == "" {
		jsonError(w, http.StatusBadRequest, "loop is required")
		return false
	}
	return true
}

func (s *Server) handleCommandStart(w http.ResponseWriter, r *http.Request) {
	s.execCommand(w, r, plc.Command{Command: plc.CommandStartPLC, Source: "web"})
}

func (s *Server) handleCommandStop(w http.ResponseWriter, r *http.Request) {
	s.execCommand(w, r, plc.Command{Command: plc.CommandStopPLC, Source: "web"})
}

type setpointRequest struct {
	Loop     string  `json:"loop"`
	Setpoint float64 `json:"setpoint"`
}

func (s *Server) handleCommandSetpoint(w http.ResponseWriter, r *http.Request) {
	var req setpointRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command: plc.CommandSetSetpoint,
		Loop:    req.Loop,
		Value:   &req.Setpoint,
		Source:  "web",
	})
}

type pidGainsRequest struct {
	Loop string  `json:"loop"`
	Kp   float64 `json:"kp"`
	Ki   float64 `json:"ki"`
	Kd   float64 `json:"kd"`
}

func (s *Server) handleCommandPIDGains(w http.ResponseWriter, r *http.Request) {
	var req pidGainsRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command: plc.CommandSetPIDGains,
		Loop:    req.Loop,
		Kp:      &req.Kp,
		Ki:      &req.Ki,
		Kd:      &req.Kd,
		Source:  "web",
	})
}

type modeRequest struct {
	Loop         string   `json:"loop"`
	Mode         string   `json:"mode"`
	ManualOutput *float64 `json:"manual_output,omitempty"`
}

func (s *Server) handleCommandMode(w http.ResponseWriter, r *http.Request) {
	var req modeRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command:      plc.CommandSetMode,
		Loop:         req.Loop,
		Mode:         plc.LoopMode(req.Mode),
		ManualOutput: req.ManualOutput,
		Source:       "web",
	})
}

type manualOutputRequest struct {
	Loop  string  `json:"loop"`
	Value float64 `json:"value"`
}

func (s *Server) handleCommandManualOutput(w http.ResponseWriter, r *http.Request) {
	var req manualOutputRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command: plc.CommandSetManualOutput,
		Loop:    req.Loop,
		Value:   &req.Value,
		Source:  "web",
	})
}

type injectDisturbanceRequest struct {
	Loop            string   `json:"loop"`
	Amplitude       float64  `json:"amplitude"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
}

func (s *Server) handleCommandInjectDisturbance(w http.ResponseWriter, r *http.Request) {
	var req injectDisturbanceRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command:         plc.CommandInjectDisturbance,
		Loop:            req.Loop,
		Value:           &req.Amplitude,
		DurationSeconds: req.DurationSeconds,
		Source:          "web",
	})
}

type resetLoopRequest struct {
	Loop string `json:"loop"`
}

func (s *Server) handleCommandResetLoop(w http.ResponseWriter, r *http.Request) {
	var req resetLoopRequest
	if err := decodeJSON(r, &req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !requireLoop(w, req.Loop) {
		return
	}
	s.execCommand(w, r, plc.Command{
		Command: plc.CommandResetLoop,
		Loop:    req.Loop,
		Source:  "web",
	})
}

func (s *Server) execCommand(w http.ResponseWriter, r *http.Request, cmd plc.Command) {
	cmd.ReceivedAt = time.Now().UTC()
	event, err := s.deps.CommandHandler(r.Context(), cmd)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, event)
}
