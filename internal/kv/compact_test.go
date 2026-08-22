package kv

import (
	"path/filepath"
	"testing"
	"time"
)

// TestCompactRefreshesTTLTracking ensures that Compact drops expired keys and
// that the TTL monitoring snapshot no longer reports them afterwards. Before
// the fix, Compact rebuilt the in-memory index but left the stale expiry set
// in the TTL manager, so TTLSnapshot kept listing keys that no longer existed.
func TestCompactRefreshesTTLTracking(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "kv.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// A live key with a future TTL — must remain tracked after compaction.
	if err := s.PutWithTTL("live-ttl", []byte("v"), time.Hour); err != nil {
		t.Fatal(err)
	}
	// A permanent key with no TTL — never tracked.
	if err := s.Put("permanent", []byte("v")); err != nil {
		t.Fatal(err)
	}
	// Stop the background sweeper so the next key is not auto-evicted before
	// compaction runs. This deterministically reproduces the "expired but
	// still tracked" state that compaction must reconcile.
	s.ttl.Stop()
	if err := s.PutWithTTL("stale-ttl", []byte("v"), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond) // let stale-ttl lapse past its expiry

	if err := s.Compact(); err != nil {
		t.Fatal(err)
	}

	tracked := s.TTLSnapshot()
	if len(tracked) != 1 || tracked[0].Key != "live-ttl" {
		t.Fatalf("TTLSnapshot = %+v, want [live-ttl]", tracked)
	}
	if _, err := s.Get("stale-ttl"); err == nil {
		t.Fatal("stale-ttl still readable after compaction")
	}
}
