package web

import (
	"net/http"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	jsonOK(w, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	snap := s.deps.Runtime.Snapshot()
	if snap.PLC.State != plc.StateRunning {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"not_ready","state":"` + string(snap.PLC.State) + `"}`))
		return
	}
	jsonOK(w, map[string]string{"status": "ready", "state": string(snap.PLC.State)})
}
