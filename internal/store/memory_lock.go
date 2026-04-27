package store

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

const memoryStoreWriteLockWeight int64 = 1 << 30

type contextRWMutex struct {
	once sync.Once
	sem  *semaphore.Weighted
}

func (m *contextRWMutex) Lock() {
	if err := m.LockContext(context.Background()); err != nil {
		panic(err)
	}
}

func (m *contextRWMutex) LockContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return m.semaphore().Acquire(ctx, memoryStoreWriteLockWeight)
}

func (m *contextRWMutex) Unlock() {
	m.semaphore().Release(memoryStoreWriteLockWeight)
}

func (m *contextRWMutex) RLock() {
	if err := m.RLockContext(context.Background()); err != nil {
		panic(err)
	}
}

func (m *contextRWMutex) RLockContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return m.semaphore().Acquire(ctx, 1)
}

func (m *contextRWMutex) RUnlock() {
	m.semaphore().Release(1)
}

func (m *contextRWMutex) semaphore() *semaphore.Weighted {
	m.once.Do(func() {
		m.sem = semaphore.NewWeighted(memoryStoreWriteLockWeight)
	})
	return m.sem
}
