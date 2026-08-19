package kv

import (
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"sort"
	"strings"
	"time"
)

type KeyInfo struct {
	Key       string
	Size      int
	ExpiresAt int64
}

func (s *Store) Inspect(prefix string) ([]KeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	out := make([]KeyInfo, 0)
	for key, record := range s.idx {
		if !strings.HasPrefix(key, prefix) || !s.isLive(record, now) {
			continue
		}
		out = append(out, KeyInfo{Key: key, Size: len(record.value), ExpiresAt: record.expiresAt})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}
