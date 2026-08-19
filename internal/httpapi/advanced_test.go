package httpapi

import (
	"github.com/chengjie/bytedance/logkv/internal/kv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdvancedBatchEndpoint(t *testing.T) {
	s, err := kv.Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	r := httptest.NewRequest(http.MethodPost, "/kv/apply-batch", strings.NewReader(`[{"Key":"a","Value":"b25l"}]`))
	w := httptest.NewRecorder()
	NewServer(s).Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if _, err := s.Get("a"); err != nil {
		t.Fatal(err)
	}
}
