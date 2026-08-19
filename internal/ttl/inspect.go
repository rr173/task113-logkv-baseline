package ttl

import "time"

type Expiry struct {
	Key string
	At  int64
}

func (m *Manager) Snapshot() []Expiry {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Expiry, 0, len(m.expiries))
	for key, at := range m.expiries {
		out = append(out, Expiry{Key: key, At: at})
	}
	return out
}

func (m *Manager) Due(now time.Time) []Expiry {
	items := m.Snapshot()
	out := make([]Expiry, 0)
	for _, item := range items {
		if item.At <= now.UnixNano() {
			out = append(out, item)
		}
	}
	return out
}
