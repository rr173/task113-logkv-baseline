package kv

import (
	"context"
	"github.com/chengjie/bytedance/logkv/internal/codec"
)

type Checkpoint struct {
	Records int
	Bytes   int
	Data    []byte
}

func (s *Store) Checkpoint(ctx context.Context) (Checkpoint, error) {
	items, err := s.Query(Query{})
	if err != nil {
		return Checkpoint{}, err
	}
	if err := ctx.Err(); err != nil {
		return Checkpoint{}, err
	}
	records := make([]codec.Record, 0, len(items))
	for _, item := range items {
		records = append(records, codec.Record{Key: item.Key, Value: item.Value, ExpiresAt: item.ExpiresAt})
	}
	data, err := codec.EncodeRecords(records)
	if err != nil {
		return Checkpoint{}, err
	}
	return Checkpoint{Records: len(records), Bytes: len(data), Data: data}, nil
}
