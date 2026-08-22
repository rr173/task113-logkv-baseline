// Package ttl implements a time-to-live tracker with a background sweeper that
// evicts expired keys through a pluggable store callback.
package ttl

import (
	"sync"
	"time"
)

// Store is the minimal surface the Manager needs to evict expired keys.
type Store interface {
	// DeleteExpired removes the key if and only if it is currently present,
	// live (not deleted) and past its expiry. It returns true when a key was
	// actually removed.
	DeleteExpired(key string) (bool, error)
}

// Manager tracks per-key expiry timestamps and evicts them on a ticker.
type Manager struct {
	mu       sync.Mutex
	expiries map[string]int64
	store    Store
	interval time.Duration
	stop     chan struct{}
	wg       sync.WaitGroup
	running  bool
}

// NewManager creates a Manager bound to store. The sweeper interval defaults
// to 200ms when zero.
func NewManager(store Store) *Manager {
	return &Manager{
		expiries: make(map[string]int64),
		store:    store,
		interval: 200 * time.Millisecond,
		stop:     make(chan struct{}),
	}
}

// Schedule records that key expires at the given unix-nanosecond timestamp.
func (m *Manager) Schedule(key string, at int64) {
	m.mu.Lock()
	if _, exists := m.expiries[key]; exists {
		m.mu.Unlock()
		return
	}
	m.expiries[key] = at
	m.mu.Unlock()
}

// Unschedule removes any pending expiry for key.
func (m *Manager) Unschedule(key string) {
	m.mu.Lock()
	delete(m.expiries, key)
	m.mu.Unlock()
}

// Reset drops every tracked expiry. Callers that wholesale replace the keyset
// (such as Store.Restore) use it so the pending expiry plan matches only the
// new records, instead of retaining stale plans for replaced or dropped keys.
func (m *Manager) Reset() {
	m.mu.Lock()
	m.expiries = make(map[string]int64)
	m.mu.Unlock()
}

// Len returns the number of tracked keys.
func (m *Manager) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.expiries)
}

// Start launches the background sweeper.
func (m *Manager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-m.stop:
				return
			case <-ticker.C:
				m.sweep()
			}
		}
	}()
}

// Stop terminates the sweeper and waits for it to exit.
func (m *Manager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	close(m.stop)
	m.wg.Wait()
}

// SweepOnce performs a single eviction pass and returns the number of keys
// removed. It is safe to call manually for tests.
func (m *Manager) SweepOnce() int {
	return m.sweep()
}

func (m *Manager) sweep() int {
	m.mu.Lock()
	now := time.Now().UnixNano()
	var due []string
	for k, at := range m.expiries {
		if at <= now {
			due = append(due, k)
		}
	}
	m.mu.Unlock()

	removed := 0
	for _, k := range due {
		ok, _ := m.store.DeleteExpired(k)
		if ok {
			removed++
		}
		m.Unschedule(k)
	}
	return removed
}
