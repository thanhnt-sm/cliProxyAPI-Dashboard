package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// Manager manages rate limiters for different keys.
type Manager struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
}

// NewManager creates a new rate limiter manager.
func NewManager() *Manager {
	return &Manager{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter returns or creates a limiter for the given key with the specified limit (requests per minute).
func (m *Manager) GetLimiter(key string, limitRPM int) *rate.Limiter {
	m.mu.RLock()
	limiter, exists := m.limiters[key]
	m.mu.RUnlock()

	if exists {
		// If the limit has changed, we might want to update it, but for simplicity we assume it's constant per key
		// or we can recreate it. For now, strict existing check.
		// To support dynamic updates, we should check limit.
		if limiter.Limit() == rate.Limit(limitRPM)/60.0 {
			return limiter
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if limiter, exists = m.limiters[key]; exists {
		if limiter.Limit() == rate.Limit(limitRPM)/60.0 {
			return limiter
		}
		// Update limit if changed
		limiter.SetLimit(rate.Limit(limitRPM) / 60.0)
		limiter.SetBurst(limitRPM) // Burst usually equal to limit or smaller. Let's allowing bursting up to the minute limit.
		return limiter
	}

	// Create new limiter
	// Rate is events per second. RPM / 60.
	r := rate.Limit(float64(limitRPM) / 60.0)
	// Burst size: Allow short bursts, e.g. 10% of RPM or at least 1?
	// If RPM is 60 (1/sec), burst 1 is strict.
	// Let's set burst to equal RPM to allow flexibility within the minute window? 
	// Or maybe smaller. Standard token bucket often uses burst = rate * window.
	// Users usually expect "60 requests per minute" to mean they can do 60 requests instantly and then wait.
	// So burst = limitRPM is appropriate.
	newLimiter := rate.NewLimiter(r, limitRPM) 
	m.limiters[key] = newLimiter
	
	return newLimiter
}

// Allow checks if a request is allowed for the key.
func (m *Manager) Allow(key string, limitRPM int) bool {
	if limitRPM <= 0 {
		return true // No limit
	}
	limiter := m.GetLimiter(key, limitRPM)
	return limiter.Allow()
}

// Global instance to be used by middleware
var GlobalManager = NewManager()
