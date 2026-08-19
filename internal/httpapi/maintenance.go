package httpapi

import "net/http"

func (s *Server) manifest(w http.ResponseWriter, r *http.Request) {
	checkpoint, err := s.store.Checkpoint(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"records": checkpoint.Records, "bytes": checkpoint.Bytes})
}
