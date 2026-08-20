package kv

import (
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"sort"
	"strings"
	"time"
)

type Query struct {
	Prefix   string
	Contains string
	MinSize  int
	MaxSize  int
}

func (s *Store) Query(q Query) ([]KeyValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, errors.ErrClosed
	}
	q = NormalizeQuery(q)
	now := time.Now().UnixNano()
	out := make([]KeyValue, 0)
	for key, record := range s.idx {
		if !s.isLive(record, now) || !strings.HasPrefix(key, q.Prefix) || q.Contains != "" && !strings.Contains(key, q.Contains) {
			continue
		}
		size := len(record.value)
		if q.MinSize > 0 && size < q.MinSize || q.MaxSize > 0 && size > q.MaxSize {
			continue
		}
		out = append(out, KeyValue{Key: key, Value: append([]byte(nil), record.value...), ExpiresAt: record.expiresAt})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}
