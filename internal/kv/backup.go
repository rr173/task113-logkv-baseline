package kv

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/codec"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	bolt "go.etcd.io/bbolt"
)

// Backup writes a consistent snapshot of all live records to w.
func (s *Store) Backup(w io.Writer) error {
	if err := s.Export(context.Background(), w); err != nil {
		return fmt.Errorf("backup: %v", err)
	}
	return nil
}

// Export streams all live records to w, honouring ctx cancellation between
// records.
func (s *Store) Export(ctx context.Context, w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return errors.ErrClosed
	}
	enc := codec.NewEncoder(w)
	now := time.Now().UnixNano()
	for k, r := range s.idx {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("export cancelled: %v", err)
		}
		if !s.isLive(r, now) {
			continue
		}
		if err := enc.Write(codec.Record{Key: k, Value: append([]byte(nil), r.value...), ExpiresAt: r.expiresAt}); err != nil {
			return fmt.Errorf("export write %s: %w", k, err)
		}
	}
	return nil
}

// Restore loads records from r, replacing the in-memory index and persisting
// them atomically.
func (s *Store) Restore(r io.Reader) error {
	dec := codec.NewDecoder(r)
	loaded := make(map[string]record)
	for {
		rec, err := dec.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		loaded[rec.Key] = record{value: append([]byte(nil), rec.Value...), expiresAt: rec.ExpiresAt, deleted: rec.Deleted}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.ErrClosed
	}
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		cur := b.Cursor()
		for key, _ := cur.First(); key != nil; key, _ = cur.Next() {
			if e := cur.Delete(); e != nil {
				return e
			}
		}
		for k, r := range loaded {
			if e := b.Put([]byte(k), marshalRecord(r)); e != nil {
				return e
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// The on-disk contents were wholesale replaced, so the previous TTL plan
	// must not survive: Schedule is add-if-absent and would otherwise leave a
	// replaced key on its stale (often earlier) expiry, and keys dropped by the
	// restore would linger as orphaned plans. Reset, then rebuild only from the
	// restored records so the plan matches the new keyset exactly.
	s.ttl.Reset()
	s.idx = make(map[string]*record, len(loaded))
	for k, r := range loaded {
		cp := r
		cp.value = append([]byte(nil), r.value...)
		s.idx[k] = &cp
		if r.expiresAt > 0 && !r.deleted {
			s.ttl.Schedule(k, r.expiresAt)
		}
	}
	return nil
}
