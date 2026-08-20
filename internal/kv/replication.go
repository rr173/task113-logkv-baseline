package kv

import (
	"context"
	"github.com/chengjie/bytedance/logkv/internal/codec"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"io"
	"time"
)

type ReplicationReport struct {
	Read    int
	Applied int
	Skipped int
}

func (s *Store) Replicate(ctx context.Context, r io.Reader) (ReplicationReport, error) {
	dec := codec.NewDecoder(r)
	out := ReplicationReport{}
	for {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		record, err := dec.Read()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		out.Read++
		if record.Deleted {
			if err := s.Delete(record.Key); err != nil && err != errors.ErrNotFound {
				return out, err
			}
			continue
		}
		// Skip records that are already expired on the remote: writing them
		// would make the stale value briefly visible in the local index and
		// surface as ErrTTLExpired instead of an absent key.
		now := time.Now().UnixNano()
		if record.ExpiresAt > 0 && record.ExpiresAt <= now {
			out.Skipped++
			continue
		}
		if ok, err := s.Has(record.Key); err != nil {
			return out, err
		} else if ok {
			out.Skipped++
			continue
		}
		if err := s.PutWithTTL(record.Key, record.Value, EffectiveTTL(record.ExpiresAt, time.Now())); err != nil {
			return out, err
		}
		out.Applied++
	}
}
