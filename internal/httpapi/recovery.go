package httpapi

import "net/http"

func (s *Server) recovery(w http.ResponseWriter, r *http.Request) {
	report, err := s.store.RecoveryReport()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, report)
}
