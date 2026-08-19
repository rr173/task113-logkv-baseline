package kv

import "github.com/chengjie/bytedance/logkv/internal/errors"

type Diff struct {
	Key    string
	Kind   string
	Before []byte
	After  []byte
}

func (s *Store) Diff(other *Store) ([]Diff, error) {
	left, err := s.Query(Query{})
	if err != nil {
		return nil, err
	}
	right, err := other.Query(Query{})
	if err != nil {
		return nil, err
	}
	lm, rm := map[string]KeyValue{}, map[string]KeyValue{}
	for _, item := range left {
		lm[item.Key] = item
	}
	for _, item := range right {
		rm[item.Key] = item
	}
	seen := map[string]bool{}
	out := make([]Diff, 0)
	for key, item := range lm {
		seen[key] = true
		peer, ok := rm[key]
		if !ok {
			out = append(out, Diff{Key: key, Kind: "removed", Before: item.Value})
			continue
		}
		if string(item.Value) != string(peer.Value) {
			out = append(out, Diff{Key: key, Kind: "changed", Before: item.Value, After: peer.Value})
		}
	}
	for key, item := range rm {
		if !seen[key] {
			out = append(out, Diff{Key: key, Kind: "added", After: item.Value})
		}
	}
	if len(out) == 0 && left == nil && right == nil {
		return nil, errors.ErrNotFound
	}
	return out, nil
}
