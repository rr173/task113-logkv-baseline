package kv

func NormalizeQuery(q Query) Query {
	if q.MinSize < 0 {
		q.MinSize = 0
	}
	if q.MaxSize < 0 {
		q.MaxSize = 0
	}
	// Normalize reversed bounds: when the caller writes min and max in the
	// wrong order, treat the range as [min, max] so the query returns the
	// intuitively expected set instead of an empty result.
	if q.MinSize > 0 && q.MaxSize > 0 && q.MinSize > q.MaxSize {
		q.MinSize, q.MaxSize = q.MaxSize, q.MinSize
	}
	return q
}

func (q Query) Matches(key string, size int) bool {
	if q.Prefix != "" && len(key) < len(q.Prefix) {
		return false
	}
	if q.MinSize > 0 && size < q.MinSize {
		return false
	}
	if q.MaxSize > 0 && size > q.MaxSize {
		return false
	}
	return true
}
