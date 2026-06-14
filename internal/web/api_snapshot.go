package web

import "net/http"

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.deps.Runtime.Snapshot())
}
