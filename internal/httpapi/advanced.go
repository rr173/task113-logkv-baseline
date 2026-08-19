package httpapi

import (
	"encoding/json"
	"github.com/chengjie/bytedance/logkv/internal/kv"
	"net/http"
)

func (s *Server) applyBatch(w http.ResponseWriter, r *http.Request) {
	var items []kv.BatchItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		writeErr(w, 400, err)
		return
	}
	result, err := s.store.ApplyBatch(items)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) inspect(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Inspect(r.URL.Query().Get("prefix"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, items)
}

func (s *Server) namespaces(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Namespaces(r.URL.Query().Get("separator"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, items)
}

func (s *Server) verify(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.Verify()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) merge(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.Merge(r.Body)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, result)
}

func (s *Server) ttlSnapshot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"tracked": s.store.TTLSnapshot()})
}
