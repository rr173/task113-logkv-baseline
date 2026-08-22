package kv

import (
	"time"

	"github.com/chengjie/bytedance/logkv/internal/errors"
	bolt "go.etcd.io/bbolt"
)

// Compact reclaims space by rewriting the bucket with only live records,
// discarding tombstones and expired entries. It rebuilds the in-memory index
// and the TTL tracker to match.
func (s *Store) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.ErrClosed
	}
	now := time.Now().UnixNano()
	live := make(map[string]record, len(s.idx))
	for k, r := range s.idx {
		if s.isLive(r, now) {
			cp := *r
			cp.value = append([]byte(nil), r.value...)
			live[k] = cp
		}
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		cur := b.Cursor()
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			if err := cur.Delete(); err != nil {
				return err
			}
		}
		for k, r := range live {
			if err := b.Put([]byte(k), marshalRecord(r)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Drop expiry tracking for every key that is no longer present so the TTL
	// snapshot reflects actual storage, then reschedule the surviving keys.
	for k := range s.idx {
		s.ttl.Unschedule(k)
	}
	s.idx = make(map[string]*record, len(live))
	for k, r := range live {
		cp := r
		s.idx[k] = &cp
		if r.expiresAt > 0 && !r.deleted {
			s.ttl.Schedule(k, r.expiresAt)
		}
	}
	s.compactions++
	s.lastCompact = now
	return nil
}
