package kv

import (
	"fmt"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"time"
)

type RecoveryReport struct {
	Checked int
	Live    int
	Expired int
	Deleted int
}

func (s *Store) RecoveryReport() (RecoveryReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return RecoveryReport{}, errors.ErrClosed
	}
	now := time.Now().UnixNano()
	out := RecoveryReport{}
	for _, record := range s.idx {
		out.Checked++
		if record.deleted {
			out.Deleted++
			continue
		}
		if record.expiresAt > 0 && record.expiresAt <= now {
			out.Expired++
		} else {
			out.Live++
		}
	}
	if out.Checked < out.Live+out.Expired+out.Deleted {
		return out, fmt.Errorf("recovery accounting mismatch")
	}
	return out, nil
}
