package kv

import (
	"fmt"
	"time"
)

type BatchItem struct {
	Key   string
	Value []byte
	TTL   time.Duration
}
type BatchResult struct {
	Applied   int
	FailedKey string
}

// ApplyBatch validates all keys before applying writes, so malformed input
// cannot leave a partially applied request. Existing values are replaced in
// the same durable path as PutWithTTL.
func (s *Store) ApplyBatch(items []BatchItem) (BatchResult, error) {
	for _, item := range items {
		if item.Key == "" {
			return BatchResult{FailedKey: item.Key}, fmt.Errorf("empty batch key")
		}
		if item.TTL < 0 {
			return BatchResult{FailedKey: item.Key}, fmt.Errorf("negative ttl for %s", item.Key)
		}
	}
	result := BatchResult{}
	for _, item := range items {
		if err := s.PutWithTTL(item.Key, item.Value, item.TTL); err != nil {
			result.FailedKey = item.Key
			return result, err
		}
		result.Applied++
	}
	return result, nil
}
