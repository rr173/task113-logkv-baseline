package httpapi

import (
	"github.com/chengjie/bytedance/logkv/internal/kv"
	"net/http"
	"strconv"
)

func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	q := kv.Query{Prefix: r.URL.Query().Get("prefix"), Contains: r.URL.Query().Get("contains")}
	q.MinSize, _ = strconv.Atoi(r.URL.Query().Get("min_size"))
	q.MaxSize, _ = strconv.Atoi(r.URL.Query().Get("max_size"))
	items, err := s.store.Query(q)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"count": len(items), "items": items})
}
func (s *Server) checkpoint(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.Checkpoint(r.Context())
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	result.Data = nil
	writeJSON(w, 200, result)
}
