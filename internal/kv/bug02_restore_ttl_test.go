package kv

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"
)

func TestBug02RestoreReplacesExistingTTLSchedule(t *testing.T) {
	dir := t.TempDir()
	source, err := Open(filepath.Join(dir, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := source.PutWithTTL("session", []byte("fresh"), time.Second); err != nil {
		t.Fatal(err)
	}
	var snapshot bytes.Buffer
	if err := source.Backup(&snapshot); err != nil {
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}

	target, err := Open(filepath.Join(dir, "target.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.PutWithTTL("session", []byte("stale"), 20*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := target.Restore(&snapshot); err != nil {
		t.Fatal(err)
	}
	tracked := target.TTLSnapshot()
	if len(tracked) != 1 || tracked[0].Key != "session" || tracked[0].At < time.Now().Add(500*time.Millisecond).UnixNano() {
		t.Fatalf("tracked TTLs = %+v, want restored session TTL with new deadline", tracked)
	}
}
