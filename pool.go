package main

import (
	"log"
	"sync/atomic"
)

// ServerPool holds the collection of backends and tracks the current index for load balancing
type ServerPool struct {
	backends []*Backend
	current  uint64
}

// AddBackend adds a new server to the pool
func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

// NextIndex atomically increments the counter and returns an index
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// GetNextPeer returns the next available (alive) server to route a request to
func (s *ServerPool) GetNextPeer() *Backend {
	// Loop through all backends to find one that is alive
	// This ensures we don't get stuck if the exact "next" one is down
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and iterate full length
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		// If we find an alive backend, return it immediately
		if s.backends[idx].IsAlive() {
			// update the current so the next request starts from here
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	
	// We only hit this if ALL backends are dead
	log.Println("CRITICAL: No healthy backends available!")
	return nil
}
