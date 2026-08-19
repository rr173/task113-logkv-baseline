package kv

import "github.com/chengjie/bytedance/logkv/internal/ttl"

func (s *Store) TTLSnapshot() []ttl.Expiry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.ttl == nil {
		return []ttl.Expiry{}
	}
	return s.ttl.Snapshot()
}
