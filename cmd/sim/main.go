package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// A simple backend server simulation
	var port int
	var serverName string
	flag.IntVar(&port, "port", 8081, "Port to serve")
	flag.StringVar(&serverName, "name", "Server-A", "Name of the server instance")
	flag.Parse()

	// Simple handler that responds with the server's identity
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] Received request from %s\n", serverName, r.RemoteAddr)
		
		// Add some headers to prove it came from this specific backend
		w.Header().Set("X-Backend-Server", serverName)
		w.WriteHeader(http.StatusOK)
		
		fmt.Fprintf(w, "Hello from %s!\n", serverName)
		fmt.Fprintf(w, "Request Path: %s\n", r.URL.Path)
	})

	// Add a dedicated health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	log.Printf("Starting simulated backend [%s] on port %d...\n", serverName, port)
	log.Printf("Try accessing: http://localhost:%d\n", port)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("Server [%s] stopped: %v\n", serverName, err)
		os.Exit(1)
	}
}
