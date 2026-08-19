package kv

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/codec"
	apperrors "github.com/chengjie/bytedance/logkv/internal/errors"
)

func TestBug04ReplicationSkipsExpiredRecords(t *testing.T) {
	store, err := Open(t.TempDir() + "/store.db")
	if err != nil { t.Fatal(err) }
	defer store.Close()
	data, err := codec.EncodeRecords([]codec.Record{{Key: "expired", Value: []byte("stale"), ExpiresAt: time.Now().Add(-time.Second).UnixNano()}})
	if err != nil { t.Fatal(err) }
	if _, err := store.Replicate(context.Background(), bytes.NewReader(data)); err != nil { t.Fatal(err) }
	if _, err := store.Get("expired"); err != apperrors.ErrNotFound { t.Fatalf("expired replicated value error = %v, want ErrNotFound", err) }
}
