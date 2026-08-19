package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestBug09_ReversedRangeIsRejected(t *testing.T) {
	s, err := kv.Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("m", []byte("value")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RangeScan("z", "a"); err == nil {
		t.Fatal("store accepted a reversed range")
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/kv/range?start=z&end=a", nil)
	NewServer(s).Handler().ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("reversed range status=%d body=%s, want %d", w.Code, w.Body.String(), http.StatusBadRequest)
	}
}
