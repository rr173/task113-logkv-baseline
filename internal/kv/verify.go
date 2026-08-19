package kv

import (
	"fmt"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"time"
)

type Verification struct {
	Checked    int
	Live       int
	Tombstones int
	Invalid    int
}

func (s *Store) Verify() (Verification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return Verification{}, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	out := Verification{}
	for key, record := range s.idx {
		out.Checked++
		if key == "" || record == nil {
			out.Invalid++
			continue
		}
		if record.deleted || record.expiresAt > 0 && record.expiresAt <= now {
			out.Tombstones++
		} else {
			out.Live++
		}
	}
	if out.Invalid > 0 {
		return out, fmt.Errorf("invalid index entries: %d", out.Invalid)
	}
	return out, nil
}
