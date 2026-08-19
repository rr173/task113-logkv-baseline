package kv

import (
	"bytes"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"time"
)

func (s *Store) CompareAndSwap(key string, expected, next []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false, errors.ErrClosed
	}
	current, ok := s.idx[key]
	if !ok || !s.isLive(current, time.Now().UnixNano()) {
		return false, errors.ErrNotFound
	}
	if !bytes.Equal(current.value, expected) {
		return false, nil
	}
	replacement := record{value: append([]byte(nil), next...)}
	if err := s.writeRecord(key, replacement); err != nil {
		return false, err
	}
	s.idx[key] = &replacement
	return true, nil
}
