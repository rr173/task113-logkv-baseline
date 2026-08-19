package ttl

import (
	"testing"
	"time"
)

type fakeTTLStore struct{}

func (fakeTTLStore) DeleteExpired(string) (bool, error) { return true, nil }

func TestSnapshotAndDue(t *testing.T) {
	m := NewManager(fakeTTLStore{})
	m.Schedule("old", time.Now().Add(-time.Second).UnixNano())
	m.Schedule("new", time.Now().Add(time.Hour).UnixNano())
	if len(m.Snapshot()) != 2 || len(m.Due(time.Now())) != 1 {
		t.Fatal("unexpected ttl inspection")
	}
}
