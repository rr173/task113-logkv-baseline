package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestBug10_TTLKeysExcludePersistentKeys(t *testing.T) {
	s, err := kv.Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("persistent", []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := s.PutWithTTL("expiring", []byte("value"), time.Minute); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/kv/ttl/keys", nil)
	NewServer(s).Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("ttl keys status=%d body=%s", w.Code, w.Body.String())
	}
	var out struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Keys) != 1 || out.Keys[0] != "expiring" {
		t.Fatalf("ttl keys=%v, want only expiring", out.Keys)
	}
}
