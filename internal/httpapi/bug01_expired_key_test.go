package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestBug01ExpiredKeyUsesNotFoundResponse(t *testing.T) {
	store, err := kv.Open(t.TempDir() + "/store.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.PutWithTTL("session", []byte("value"), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/kv/session", nil)
	res := httptest.NewRecorder()
	NewServer(store).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d for expired key", res.Code, http.StatusNotFound)
	}
}
