package kv

import "github.com/chengjie/bytedance/logkv/internal/errors"

type Health struct {
	Open       bool  `json:"open"`
	Keys       int   `json:"keys"`
	Bytes      int64 `json:"bytes"`
	TTLTracked int   `json:"ttl_tracked"`
}

func (s *Store) Health() (Health, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return Health{}, errors.ErrClosed
	}
	stats, err := s.Stats()
	if err != nil {
		return Health{}, err
	}
	tracked := 0
	if s.ttl != nil {
		tracked = s.ttl.Len()
	}
	return Health{Open: true, Keys: stats.Count, Bytes: stats.LiveBytes, TTLTracked: tracked}, nil
}

func (s *Store) Ready() bool { s.mu.RLock(); defer s.mu.RUnlock(); return !s.closed && s.db != nil }
