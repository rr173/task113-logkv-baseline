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
