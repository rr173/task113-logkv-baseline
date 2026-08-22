package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestTTLKeysExcludesPermanent(t *testing.T) {
	s, err := kv.Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("perm", []byte("p")); err != nil { // permanent key
		t.Fatal(err)
	}
	if err := s.PutWithTTL("temp", []byte("t"), time.Minute); err != nil { // TTL key
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/kv/ttl/keys", nil)
	w := httptest.NewRecorder()
	NewServer(s).Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Count int `json:"count"`
		Keys  []struct {
			Key       string `json:"key"`
			ExpiresAt int64  `json:"expires_at"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, w.Body.String())
	}
	if resp.Count != 1 || len(resp.Keys) != 1 || resp.Keys[0].Key != "temp" {
		t.Fatalf("want only [temp], got count=%d keys=%+v", resp.Count, resp.Keys)
	}
	if resp.Keys[0].ExpiresAt <= 0 {
		t.Fatalf("temp has no valid expiry: %+v", resp.Keys[0])
	}
}
