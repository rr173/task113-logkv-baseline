package kv

import (
	"github.com/chengjie/bytedance/logkv/internal/codec"
	"github.com/chengjie/bytedance/logkv/internal/errors"
	"io"
	"time"
)

type MergeResult struct {
	Read    int
	Written int
}

func (s *Store) Merge(r io.Reader) (MergeResult, error) {
	dec := codec.NewDecoder(r)
	out := MergeResult{}
	for {
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
		// Skip records that are already expired by the time we merge them.
		// Otherwise PutWithTTL receives a non-positive (negative) duration,
		// clears the expiry, and writes the record as permanent — reviving
		// data that should have disappeared.
		if record.ExpiresAt > 0 && record.ExpiresAt <= time.Now().UnixNano() {
			continue
		}
		var ttlDur time.Duration
		if record.ExpiresAt > 0 {
			ttlDur = time.Until(time.Unix(0, record.ExpiresAt))
		}
		if err := s.PutWithTTL(record.Key, record.Value, ttlDur); err != nil {
			return out, err
		}
		out.Written++
	}
}
