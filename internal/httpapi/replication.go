package httpapi

import "net/http"

func (s *Server) replicate(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.Replicate(r.Context(), r.Body)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, result)
}
func (s *Server) evictions(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.EvictionCandidates(100)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, rows)
}
