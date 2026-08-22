package kv

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	apperrors "github.com/chengjie/bytedance/logkv/internal/errors"
)

func TestRestoreReplacesContentsAndPreservesValues(t *testing.T) {
	source, err := Open(filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err := source.Put("fresh", []byte("value")); err != nil {
		t.Fatal(err)
	}
	var backup bytes.Buffer
	if err := source.Backup(&backup); err != nil {
		t.Fatal(err)
	}

	destination, err := Open(filepath.Join(t.TempDir(), "destination.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	if err := destination.Put("stale", []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := destination.Restore(&backup); err != nil {
		t.Fatal(err)
	}
	got, err := destination.Get("fresh")
	if err != nil || string(got) != "value" {
		t.Fatalf("restored value = %q, %v; want value, nil", got, err)
	}
	if _, err := destination.Get("stale"); err != apperrors.ErrNotFound {
		t.Fatalf("stale key error = %v, want ErrNotFound", err)
	}
}

func TestOpenReschedulesPersistedTTL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ttl.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PutWithTTL("ephemeral", []byte("value"), time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	tracked := reopened.TTLSnapshot()
	if len(tracked) != 1 || tracked[0].Key != "ephemeral" {
		t.Fatalf("tracked TTLs = %+v, want ephemeral", tracked)
	}
}

func TestRestoreRebuildsTTLPlanToMatchRestoredKeys(t *testing.T) {
	destination, err := Open(filepath.Join(t.TempDir(), "destination.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	// Stale state before restore: "kept" carries a soon-ish expiry and "dropped"
	// a far-future one. Both pre-exist in the TTL plan.
	staleExpiry := time.Now().Add(2 * time.Second)
	if err := destination.PutWithTTL("kept", []byte("old"), time.Until(staleExpiry)); err != nil {
		t.Fatal(err)
	}
	if err := destination.PutWithTTL("dropped", []byte("old"), time.Hour); err != nil {
		t.Fatal(err)
	}

	var backup bytes.Buffer
	source, err := Open(filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	// "kept" is replaced with a long-lived value; "dropped" is absent from the
	// backup; "fresh" is brand new.
	if err := source.PutWithTTL("kept", []byte("new"), time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := source.Put("fresh", []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := source.Backup(&backup); err != nil {
		t.Fatal(err)
	}

	if err := destination.Restore(&backup); err != nil {
		t.Fatal(err)
	}
	tracked := destination.TTLSnapshot()
	byKey := make(map[string]int64, len(tracked))
	for _, e := range tracked {
		byKey[e.Key] = e.At
	}
	if _, ok := byKey["dropped"]; ok {
		t.Fatalf("dropped key still tracked after restore: %+v", tracked)
	}
	if _, ok := byKey["fresh"]; ok {
		t.Fatalf("fresh key without TTL should not be tracked: %+v", tracked)
	}
	keptAt, ok := byKey["kept"]
	if !ok {
		t.Fatalf("kept key not tracked after restore: %+v", tracked)
	}
	if keptAt <= staleExpiry.UnixNano() {
		t.Fatalf("kept key still on stale expiry %d (restored plan must use the restored expiry > %d)",
			keptAt, staleExpiry.UnixNano())
	}
}
