package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (l *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	l.RLock()
	count, exists := l.clients[ip]
	l.RUnlock()

	if !exists || count < l.limit {
		l.Lock()
		if !exists {
			go l.resetCount(ip)
		}
		l.clients[ip]++
		l.Unlock()
		return true, 0
	}

	return false, l.window
}

func (l *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(l.window)
	l.Lock()
	delete(l.clients, ip)
	l.Unlock()
}
