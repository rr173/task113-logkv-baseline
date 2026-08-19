package kv

import (
	"testing"
	"time"
)

func TestBug07_CompactionRemovesExpiredTTLTracking(t *testing.T) {
	s, err := Open(t.TempDir() + "/kv.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.PutWithTTL("expired", []byte("value"), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := s.Compact(); err != nil {
		t.Fatal(err)
	}
	if tracked := s.TTLSnapshot(); len(tracked) != 0 {
		t.Fatalf("compaction left stale TTL tracking: %+v", tracked)
	}
}
