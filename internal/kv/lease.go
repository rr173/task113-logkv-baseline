package kv

import (
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"github.com/chengjie/bytedance/logkv/internal/ttl"
	"time"
)

type ValueWithTTL struct {
	Value     []byte
	ExpiresAt int64
	Remaining time.Duration
}

func (s *Store) GetWithTTL(key string) (ValueWithTTL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ValueWithTTL{}, errors.ErrClosed
	}
	record, ok := s.idx[key]
	if !ok || !s.isLive(record, time.Now().UnixNano()) {
		return ValueWithTTL{}, errors.ErrNotFound
	}
	return ValueWithTTL{Value: append([]byte(nil), record.value...), ExpiresAt: record.expiresAt, Remaining: ttl.Remaining(record.expiresAt, time.Now())}, nil
}
