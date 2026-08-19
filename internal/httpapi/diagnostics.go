package httpapi

import "net/http"

func (s *Server) diagnostics(w http.ResponseWriter, r *http.Request) {
	report, err := s.store.Verify()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	stats, err := s.store.Stats()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"verification": report, "stats": stats})
}
