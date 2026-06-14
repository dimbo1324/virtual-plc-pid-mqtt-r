package web

import (
	"net/http"
	"time"
)

type statusResponse struct {
	DeviceID   string    `json:"device_id"`
	State      string    `json:"state"`
	Uptime     string    `json:"uptime"`
	ServerTime time.Time `json:"server_time"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	snap := s.deps.Runtime.Snapshot()
	jsonOK(w, statusResponse{
		DeviceID:   snap.DeviceID,
		State:      string(snap.PLC.State),
		Uptime:     time.Since(s.started).Truncate(time.Second).String(),
		ServerTime: time.Now().UTC(),
	})
}
