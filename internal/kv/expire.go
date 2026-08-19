package kv

import "github.com/chengjie/bytedance/logkv/internal/errors"

// ExpireNow performs a synchronous sweep of expired keys and returns the number
// of keys removed.
func (s *Store) ExpireNow() (int, error) {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed {
		return 0, errors.ErrClosed
	}
	return s.ttl.SweepOnce(), nil
}
