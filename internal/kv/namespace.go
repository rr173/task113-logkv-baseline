package kv

import "strings"

type NamespaceStat struct {
	Namespace string
	Keys      int
	Bytes     int64
}

func (s *Store) Namespaces(separator string) ([]NamespaceStat, error) {
	if separator == "" {
		separator = "/"
	}
	items, err := s.Inspect("")
	if err != nil {
		return nil, err
	}
	groups := map[string]*NamespaceStat{}
	for _, item := range items {
		name := item.Key
		if idx := strings.Index(name, separator); idx >= 0 {
			name = name[:idx]
		}
		group := groups[name]
		if group == nil {
			group = &NamespaceStat{Namespace: name}
			groups[name] = group
		}
		group.Keys++
		group.Bytes += int64(item.Size)
	}
	out := make([]NamespaceStat, 0, len(groups))
	for _, group := range groups {
		out = append(out, *group)
	}
	return out, nil
}
