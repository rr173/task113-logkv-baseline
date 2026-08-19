package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func TestBug06EmptyRangeUsesJSONArray(t *testing.T) {
	store, err := kv.Open(t.TempDir() + "/store.db")
	if err != nil { t.Fatal(err) }
	defer store.Close()
	response := httptest.NewRecorder()
	NewServer(store).Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/kv/range?start=a&end=z", nil))
	if got := response.Body.String(); got != "{\"count\":0,\"items\":[]}\n" { t.Fatalf("response = %q, want empty items array", got) }
}
