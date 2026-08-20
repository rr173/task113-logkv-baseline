// Package httpapi exposes the logkv store over HTTP.
package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/errors"
	"github.com/chengjie/bytedance/logkv/internal/kv"
)

// Server wraps a kv.Store with an HTTP handler.
type Server struct {
	store *kv.Store
	mux   *http.ServeMux
}

// NewServer builds an HTTP server for the given store.
func NewServer(store *kv.Store) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the underlying HTTP handler.
func (s *Server) Handler() http.Handler { return s.mux }

// Start begins serving on addr until the process exits.
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func (s *Server) routes() {
	s.mux.HandleFunc("PUT /kv/{key}", s.put)
	s.mux.HandleFunc("GET /kv/{key}", s.get)
	s.mux.HandleFunc("DELETE /kv/{key}", s.del)
	s.mux.HandleFunc("POST /kv/{key}/ttl", s.setTTL)
	s.mux.HandleFunc("GET /kv/{key}/ttl", s.getTTL)
	s.mux.HandleFunc("POST /kv/{key}/exists", s.exists)
	s.mux.HandleFunc("POST /kv/batch", s.batch)
	s.mux.HandleFunc("POST /kv/apply-batch", s.applyBatch)
	s.mux.HandleFunc("GET /kv/range", s.rangeScan)
	s.mux.HandleFunc("GET /kv/keys", s.keys)
	s.mux.HandleFunc("GET /kv/inspect", s.inspect)
	s.mux.HandleFunc("GET /kv/query", s.query)
	s.mux.HandleFunc("GET /kv/namespaces", s.namespaces)
	s.mux.HandleFunc("GET /kv/size", s.size)
	s.mux.HandleFunc("GET /kv/ttl/keys", s.ttlKeys)
	s.mux.HandleFunc("POST /kv/compact", s.compact)
	s.mux.HandleFunc("POST /kv/expire", s.expire)
	s.mux.HandleFunc("POST /kv/flush", s.flush)
	s.mux.HandleFunc("GET /kv/stats", s.stats)
	s.mux.HandleFunc("GET /kv/verify", s.verify)
	s.mux.HandleFunc("GET /kv/checkpoint", s.checkpoint)
	s.mux.HandleFunc("GET /kv/recovery", s.recovery)
	s.mux.HandleFunc("GET /kv/manifest", s.manifest)
	s.mux.HandleFunc("GET /kv/diagnostics", s.diagnostics)
	s.mux.HandleFunc("GET /kv/export", s.exportHTTP)
	s.mux.HandleFunc("POST /kv/backup", s.backup)
	s.mux.HandleFunc("POST /kv/restore", s.restore)
	s.mux.HandleFunc("POST /kv/merge", s.merge)
	s.mux.HandleFunc("POST /kv/replicate", s.replicate)
	s.mux.HandleFunc("GET /kv/evictions", s.evictions)
	s.mux.HandleFunc("GET /kv/ttl/snapshot", s.ttlSnapshot)
	s.mux.HandleFunc("GET /healthz", s.healthz)
	s.mux.HandleFunc("GET /healthz/detail", s.healthDetail)
	s.mux.HandleFunc("GET /metrics/sizes", s.histogram)
	s.mux.HandleFunc("GET /metrics", s.metrics)
}

func (s *Server) put(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.store.Put(key, body); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "key": key})
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	val, err := s.store.Get(key)
	if err != nil {
		if err == errors.ErrNotFound {
			writeJSON(w, 404, map[string]any{"found": false, "key": key})
			return
		}
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"found": true, "key": key, "value": string(val)})
}

func (s *Server) del(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := s.store.Delete(key); err != nil {
		if err == errors.ErrNotFound {
			writeJSON(w, 404, map[string]string{"deleted": "false", "key": key})
			return
		}
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"deleted": "true", "key": key})
}

func (s *Server) setTTL(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	ttlStr := r.URL.Query().Get("seconds")
	secs, err := strconv.Atoi(ttlStr)
	if err != nil {
		writeErr(w, 400, errors.ErrInvalidKey)
		return
	}
	val, err := s.store.Get(key)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	if err := s.store.PutWithTTL(key, val, time.Duration(secs)*time.Second); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok", "key": key})
}

func (s *Server) getTTL(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	val, err := s.store.Get(key)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	_ = val
	writeJSON(w, 200, map[string]string{"key": key, "ttl": "active"})
}

func (s *Server) exists(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	ok, err := s.store.Has(key)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"exists": ok, "key_exists": ok})
}

func (s *Server) batch(w http.ResponseWriter, r *http.Request) {
	var items []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		writeErr(w, 400, err)
		return
	}
	for _, it := range items {
		if err := s.store.Put(it.Key, []byte(it.Value)); err != nil {
			writeErr(w, 500, err)
			return
		}
	}
	writeJSON(w, 200, map[string]int{"count": len(items)})
}

func (s *Server) rangeScan(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	start := q.Get("start")
	end := q.Get("end")
	if end == "" {
		end = "\xff"
	}
	res, err := s.store.RangeScan(start, end)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if res == nil {
		// Keep a stable JSON array ([]) for empty results so clients can
		// iterate items directly without a nil check.
		res = make([]kv.KeyValue, 0)
	}
	writeJSON(w, 200, map[string]any{"count": len(res), "items": res})
}

func (s *Server) keys(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	keys, err := s.store.Keys(prefix)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"count": len(keys), "keys": keys})
}

func (s *Server) size(w http.ResponseWriter, r *http.Request) {
	n, err := s.store.Len()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]int{"size": n})
}

func (s *Server) ttlKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.store.Keys("")
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"count": len(keys), "keys": keys})
}

func (s *Server) compact(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Compact(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) expire(w http.ResponseWriter, r *http.Request) {
	n, err := s.store.ExpireNow()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]int{"expired": n})
}

func (s *Server) flush(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Compact(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.Stats()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, st)
}

func (s *Server) exportHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	if err := s.store.Export(r.Context(), w); err != nil {
		writeErr(w, 500, err)
	}
}

func (s *Server) backup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	if err := s.store.Backup(w); err != nil {
		writeErr(w, 500, err)
	}
}

func (s *Server) restore(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Restore(r.Body); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	st, err := s.store.Stats()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var b strings.Builder
	b.WriteString("# logkv metrics\n")
	b.WriteString("logkv_keys " + strconv.Itoa(st.Count) + "\n")
	b.WriteString("logkv_tombstones " + strconv.Itoa(st.Tombstones) + "\n")
	b.WriteString("logkv_compactions " + strconv.FormatInt(st.Compactions, 10) + "\n")
	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.WriteString(w, b.String())
}
