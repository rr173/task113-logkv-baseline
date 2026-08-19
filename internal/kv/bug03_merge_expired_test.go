package kv

import (
	"bytes"
	"testing"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/codec"
	apperrors "github.com/chengjie/bytedance/logkv/internal/errors"
)

func TestBug03MergeSkipsExpiredRecords(t *testing.T) {
	store, err := Open(t.TempDir() + "/store.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Put("active", []byte("value")); err != nil {
		t.Fatal(err)
	}
	data, err := codec.EncodeRecords([]codec.Record{{Key: "expired", Value: []byte("stale"), ExpiresAt: time.Now().Add(-time.Second).UnixNano()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Merge(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("expired"); err != apperrors.ErrNotFound {
		t.Fatalf("expired merge value error = %v, want ErrNotFound", err)
	}
}
