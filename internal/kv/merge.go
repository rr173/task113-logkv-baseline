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
		if !ShouldRestore(record.ExpiresAt, time.Now()) {
			continue
		}
		if err := s.PutWithTTL(record.Key, record.Value, time.Until(time.Unix(0, record.ExpiresAt))); err != nil {
			return out, err
		}
		out.Written++
	}
}
