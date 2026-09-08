package storage

import (
	"context"
	"sync"
	"time"

	"watchtower/internal/checker"
)

// MemoryStore is an in-memory, thread-safe implementation of Store.
type MemoryStore struct {
	mu         sync.RWMutex
	failures   map[string]int64
	lastStatus map[string]checker.Status
	cooldowns  map[string]time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		failures:   make(map[string]int64),
		lastStatus: make(map[string]checker.Status),
		cooldowns:  make(map[string]time.Time),
	}
}

func (m *MemoryStore) IncrFailures(_ context.Context, target string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.failures[target]++
	return m.failures[target], nil
}

func (m *MemoryStore) ResetFailures(_ context.Context, target string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.failures, target)
	return nil
}

func (m *MemoryStore) GetFailures(_ context.Context, target string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.failures[target], nil
}

func (m *MemoryStore) SetCooldown(_ context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cooldowns[key] = time.Now().Add(ttl)
	return nil
}

func (m *MemoryStore) IsOnCooldown(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt, exists := m.cooldowns[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(expiresAt) {
		delete(m.cooldowns, key)
		return false, nil
	}

	return true, nil
}

func (m *MemoryStore) SetLastStatus(_ context.Context, target string, status checker.Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastStatus[target] = status
	return nil
}

func (m *MemoryStore) GetLastStatus(_ context.Context, target string) (checker.Status, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.lastStatus[target]
	if !exists {
		return "", nil
	}
	return status, nil
}

func (m *MemoryStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.failures)
	clear(m.lastStatus)
	clear(m.cooldowns)
	return nil
}
