package main

import (
	"log"
	"net"
	"net/url"
	"sync"
	"time"
)

// isBackendAlive checks if a backend is alive by attempting a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	// Dial to the host to check if it's accepting connections.
	// For a more advanced version, we'd actually send an HTTP GET to a /health endpoint,
	// but a raw TCP dial is faster and works generically for any HTTP server.
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Server [%s] is unreachable\n", u.String())
		return false
	}
	_ = conn.Close() // we just needed to see if it connects
	return true
}

// HealthCheck runs continuously to ping all backends and update their status
func (s *ServerPool) HealthCheck() {
	for {
		// We want to ping all servers concurrently to avoid blocking
		var wg sync.WaitGroup
		for _, b := range s.backends {
			wg.Add(1)
			// Launch a goroutine for each backend
			go func(backend *Backend) {
				defer wg.Done()
				
				status := "up"
				alive := isBackendAlive(backend.URL)
				
				// Log state transitions (only if it changes)
				if alive != backend.IsAlive() {
					if !alive {
						status = "down"
					}
					log.Printf("Health Check: [%s] is now %s\n", backend.URL, status)
				}
				
				backend.SetAlive(alive)
			}(b)
		}
		
		// Wait for all concurrent checks to finish before sleeping
		wg.Wait()
		
		// Sleep for a bit before checking again to avoid slamming the backends
		time.Sleep(5 * time.Second)
	}
}
