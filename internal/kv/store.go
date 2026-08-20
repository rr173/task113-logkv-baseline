// Package kv implements an embedded ordered key/value store backed by bbolt
// with an in-memory index, TTL support, compaction and backup/export.
package kv

import (
	"encoding/binary"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chengjie/bytedance/logkv/internal/errors"
	"github.com/chengjie/bytedance/logkv/internal/ttl"
	bolt "go.etcd.io/bbolt"
)

// KeyValue is a live key/value pair returned by range/scan operations.
type KeyValue struct {
	Key       string
	Value     []byte
	ExpiresAt int64
}

// Stats reports store cardinality and compaction history.
type Stats struct {
	Count       int
	LiveBytes   int64
	Tombstones  int
	Compactions int64
	LastCompact int64
}

type record struct {
	value     []byte
	expiresAt int64
	deleted   bool
}

func marshalRecord(r record) []byte {
	buf := make([]byte, 1+8+4+len(r.value))
	if r.deleted {
		buf[0] = 1
	}
	binary.BigEndian.PutUint64(buf[1:9], uint64(r.expiresAt))
	binary.BigEndian.PutUint32(buf[9:13], uint32(len(r.value)))
	copy(buf[13:], r.value)
	return buf
}

func unmarshalRecord(b []byte) (record, error) {
	if len(b) < 13 {
		return record{}, errors.ErrInvalidKey
	}
	r := record{deleted: b[0] == 1, expiresAt: int64(binary.BigEndian.Uint64(b[1:9]))}
	n := int(binary.BigEndian.Uint32(b[9:13]))
	if len(b) != 13+n {
		return record{}, errors.ErrInvalidKey
	}
	if n > 0 {
		r.value = append([]byte(nil), b[13:13+n]...)
	}
	return r, nil
}

// Store is the logkv key/value store.
type Store struct {
	mu          sync.RWMutex
	db          *bolt.DB
	idx         map[string]*record
	ttl         *ttl.Manager
	compactions int64
	lastCompact int64
	closed      bool
	path        string
}

var bucketName = []byte("kv")

// Open opens (or creates) a store at path and recovers its state.
func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0o644, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, idx: make(map[string]*record), path: path}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(bucketName)
		return e
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.recover(); err != nil {
		_ = db.Close()
		return nil, err
	}
	s.ttl = ttl.NewManager(s)
	for key, rec := range s.idx {
		if !rec.deleted && rec.expiresAt > 0 {
			s.ttl.Schedule(key, rec.expiresAt)
		}
	}
	s.ttl.Start()
	return s, nil
}

func (s *Store) recover() error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			r, err := unmarshalRecord(v)
			if err != nil {
				return err
			}
			cp := r
			cp.value = append([]byte(nil), r.value...)
			s.idx[string(k)] = &cp
			return nil
		})
	})
}

func (s *Store) writeRecord(key string, r record) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(key), marshalRecord(r))
	})
}

func (s *Store) isLive(r *record, now int64) bool {
	if r == nil || r.deleted {
		return false
	}
	if r.expiresAt > 0 && r.expiresAt <= now {
		return false
	}
	return true
}

// Put stores value under key, replacing any existing entry and clearing TTL.
func (s *Store) Put(key string, value []byte) error {
	return s.PutWithTTL(key, value, 0)
}

// PutWithTTL stores value under key with an optional TTL. A non-positive TTL
// clears any expiry.
func (s *Store) PutWithTTL(key string, value []byte, ttlDur time.Duration) error {
	if key == "" {
		return errors.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.ErrClosed
	}
	var expiresAt int64
	if ttlDur > 0 {
		expiresAt = time.Now().Add(ttlDur).UnixNano()
	}
	rec := &record{value: append([]byte(nil), value...), expiresAt: expiresAt}
	if err := s.writeRecord(key, *rec); err != nil {
		return err
	}
	s.idx[key] = rec
	if ttlDur > 0 {
		s.ttl.Schedule(key, expiresAt)
	} else {
		s.ttl.Unschedule(key)
	}
	return nil
}

// PutIfAbsent stores value only when key is absent. It returns
// errors.ErrAlreadyExists when the key already exists and is live.
func (s *Store) PutIfAbsent(key string, value []byte) error {
	if key == "" {
		return errors.ErrInvalidKey
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.ErrClosed
	}
	if r, ok := s.idx[key]; ok && s.isLive(r, time.Now().UnixNano()) {
		return errors.ErrAlreadyExists
	}
	rec := &record{value: append([]byte(nil), value...)}
	if err := s.writeRecord(key, *rec); err != nil {
		return err
	}
	s.idx[key] = rec
	s.ttl.Unschedule(key)
	return nil
}

// Get returns the value stored under key. It returns errors.ErrNotFound when
// the key is absent, deleted or expired.
func (s *Store) Get(key string) ([]byte, error) {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil, errors.ErrClosed
	}
	r, ok := s.idx[key]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.ErrNotFound
	}
	if r.deleted {
		return nil, errors.ErrNotFound
	}
	if r.expiresAt > 0 && r.expiresAt <= time.Now().UnixNano() {
		return nil, errors.ErrTTLExpired
	}
	return append([]byte(nil), r.value...), nil
}

// Has reports whether key is present, live and unexpired.
func (s *Store) Has(key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return false, errors.ErrClosed
	}
	r, ok := s.idx[key]
	if !ok {
		return false, nil
	}
	return s.isLive(r, time.Now().UnixNano()), nil
}

// Delete removes key. Deleting a missing key returns errors.ErrNotFound.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.ErrClosed
	}
	r, ok := s.idx[key]
	if !ok {
		return errors.ErrNotFound
	}
	r.deleted = true
	r.value = nil
	r.expiresAt = 0
	if err := s.writeRecord(key, *r); err != nil {
		return err
	}
	s.ttl.Unschedule(key)
	return nil
}

// DeleteExpired implements ttl.Store. It removes key only when it is present,
// not already tombstoned, and past its expiry.
func (s *Store) DeleteExpired(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false, nil
	}
	r, ok := s.idx[key]
	if !ok {
		return false, nil
	}
	if r.deleted {
		return false, nil
	}
	now := time.Now().UnixNano()
	if r.expiresAt == 0 || r.expiresAt > now {
		return false, nil
	}
	r.deleted = true
	r.value = nil
	r.expiresAt = 0
	if err := s.writeRecord(key, *r); err != nil {
		return false, err
	}
	s.ttl.Unschedule(key)
	return true, nil
}

// Len returns the number of live keys.
func (s *Store) Len() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	n := 0
	for _, r := range s.idx {
		if s.isLive(r, now) {
			n++
		}
	}
	return n, nil
}

// Keys returns live keys matching prefix, sorted lexicographically.
func (s *Store) Keys(prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	var out []string
	for k, r := range s.idx {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if s.isLive(r, now) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out, nil
}

// RangeScan returns live key/value pairs with start <= key < end, sorted by
// key. An empty result is returned as a non-nil empty slice. It returns
// ErrInvalidRange when start sorts after end, so callers can distinguish a
// reversed range from a genuinely empty one.
func (s *Store) RangeScan(start, end string) ([]KeyValue, error) {
	if start > end {
		return nil, errors.ErrInvalidRange
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	var keys []string
	for k, r := range s.idx {
		if !s.isLive(r, now) {
			continue
		}
		if k >= start && k < end {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return []KeyValue{}, nil
	}
	sort.Strings(keys)
	out := make([]KeyValue, 0, len(keys))
	for _, k := range keys {
		r := s.idx[k]
		out = append(out, KeyValue{Key: k, Value: append([]byte(nil), r.value...), ExpiresAt: r.expiresAt})
	}
	return out, nil
}

// Stats returns a snapshot of store cardinality and compaction history.
func (s *Store) Stats() (Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return Stats{}, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	var count, tombs int
	var liveBytes int64
	for _, r := range s.idx {
		if r.deleted {
			tombs++
			continue
		}
		if r.expiresAt > 0 && r.expiresAt <= now {
			tombs++
			continue
		}
		count++
		liveBytes += int64(len(r.value))
	}
	return Stats{
		Count:       count,
		LiveBytes:   liveBytes,
		Tombstones:  tombs,
		Compactions: s.compactions,
		LastCompact: s.lastCompact,
	}, nil
}

// Close stops the sweeper and closes the database.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.ttl.Stop()
	return s.db.Close()
}

// Path returns the database file path.
func (s *Store) Path() string { return s.path }
