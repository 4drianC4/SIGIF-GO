package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
	NowUTC() time.Time
	UnixMilli() int64
}

type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) NowUTC() time.Time {
	return time.Now().UTC()
}

func (RealClock) UnixMilli() int64 {
	return time.Now().UnixMilli()
}

type MockClock struct {
	mu   sync.Mutex
	time time.Time
}

func NewMockClock(t time.Time) *MockClock {
	return &MockClock{time: t}
}

func (m *MockClock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.time
}

func (m *MockClock) NowUTC() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.time.UTC()
}

func (m *MockClock) UnixMilli() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.time.UnixMilli()
}

func (m *MockClock) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.time = t
}

func (m *MockClock) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.time = m.time.Add(d)
}