package httpapi

import "net/http"

func (s *Server) healthDetail(w http.ResponseWriter, r *http.Request) {
	health, err := s.store.Health()
	if err != nil {
		writeErr(w, 503, err)
		return
	}
	writeJSON(w, 200, health)
}

func (s *Server) histogram(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.SizeHistogram()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, result)
}
