package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

var ErrOpen = errors.New("circuit breaker open")

type Breaker struct {
	mu          sync.Mutex
	state       state
	failures    int
	successes   int
	lastFailure time.Time
	threshold   int
	timeout     time.Duration
	halfOpenMax int
}

type Config struct {
	Threshold   int
	Timeout     time.Duration
	HalfOpenMax int
}

func New(cfg Config) *Breaker {
	return &Breaker{
		threshold:   cfg.Threshold,
		timeout:     cfg.Timeout,
		halfOpenMax: cfg.HalfOpenMax,
	}
}

func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case stateClosed:
		return nil

	case stateOpen:
		if time.Since(b.lastFailure) >= b.timeout {
			b.state = stateHalfOpen
			b.successes = 0
			return nil
		}
		return ErrOpen

	case stateHalfOpen:
		if b.successes < b.halfOpenMax {
			return nil
		}
		return ErrOpen
	}

	return nil
}

func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case stateClosed:
		b.failures = 0
	case stateHalfOpen:
		b.successes++
		if b.successes >= b.halfOpenMax {
			b.state = stateClosed
			b.failures = 0
		}
	}
}

func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastFailure = time.Now()
	switch b.state {
	case stateClosed:
		b.failures++
		if b.failures >= b.threshold {
			b.state = stateOpen
		}
	case stateHalfOpen:
		b.state = stateOpen
	}
}

func (b *Breaker) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state == stateOpen
}
