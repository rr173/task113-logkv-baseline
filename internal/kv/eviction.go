package kv

import (
	"sort"
	"time"
)

type EvictionCandidate struct {
	Key       string
	Size      int
	ExpiresAt int64
}

func (s *Store) EvictionCandidates(limit int) ([]EvictionCandidate, error) {
	items, err := s.Inspect("")
	if err != nil {
		return nil, err
	}
	out := make([]EvictionCandidate, 0, len(items))
	for _, item := range items {
		out = append(out, EvictionCandidate{Key: item.Key, Size: item.Size, ExpiresAt: item.ExpiresAt})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ExpiresAt == 0 {
			return false
		}
		if out[j].ExpiresAt == 0 {
			return true
		}
		return out[i].ExpiresAt < out[j].ExpiresAt
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	_ = time.Now()
	return out, nil
}
