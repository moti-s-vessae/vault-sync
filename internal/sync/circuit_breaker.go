package sync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type cbState int

const (
	stateClosed cbState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker wraps a SecretsLoader and trips open after a threshold of
// consecutive failures, allowing recovery after a cooldown period.
type CircuitBreaker struct {
	inner      SecretsLoader
	threshold  int
	cooldown   time.Duration

	mu         sync.Mutex
	state      cbState
	failures   int
	openedAt   time.Time
}

// NewCircuitBreaker creates a CircuitBreaker wrapping inner.
// threshold is the number of consecutive failures before opening.
// cooldown is how long to wait before moving to half-open.
func NewCircuitBreaker(inner SecretsLoader, threshold int, cooldown time.Duration) (*CircuitBreaker, error) {
	if inner == nil {
		return nil, errors.New("circuit breaker: inner loader must not be nil")
	}
	if threshold < 1 {
		return nil, fmt.Errorf("circuit breaker: threshold must be >= 1, got %d", threshold)
	}
	if cooldown <= 0 {
		return nil, fmt.Errorf("circuit breaker: cooldown must be positive, got %s", cooldown)
	}
	return &CircuitBreaker{
		inner:     inner,
		threshold: threshold,
		cooldown:  cooldown,
		state:     stateClosed,
	}, nil
}

// Load delegates to the inner loader unless the circuit is open.
func (cb *CircuitBreaker) Load(ctx context.Context, path string) (map[string]string, error) {
	cb.mu.Lock()
	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = stateHalfOpen
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	case stateHalfOpen, stateClosed:
		// proceed
	}
	cb.mu.Unlock()

	secrets, err := cb.inner.Load(ctx, path)

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold || cb.state == stateHalfOpen {
			cb.state = stateOpen
			cb.openedAt = time.Now()
		}
		return nil, err
	}
	cb.failures = 0
	cb.state = stateClosed
	return secrets, nil
}

// State returns the current state label for observability.
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}
