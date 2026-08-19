package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestBug08_BackupHonorsCancelledRequest(t *testing.T) {
	s, err := kv.Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("key", []byte("value")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest(http.MethodPost, "/kv/backup", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	NewServer(s).Handler().ServeHTTP(w, r)
	if w.Code != http.StatusRequestTimeout {
		t.Fatalf("cancelled backup status=%d body=%s, want %d", w.Code, w.Body.String(), http.StatusRequestTimeout)
	}
}
