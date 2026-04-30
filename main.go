package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Global server pool
var serverPool ServerPool

// loadBalance handles incoming requests and routes them using the ServerPool
func loadBalance(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		// peer.ReverseProxy uses Director to modify the request,
		// and then sends it over to the backend.
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	
	// If no peer is returned, the system is totally down
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {
	// Parse backend server list from command line
	// e.g., --backends=http://localhost:8081,http://localhost:8082
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 8080, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance. Example: --backends=http://localhost:8081,http://localhost:8082")
	}

	// Parse servers and initialize the pool
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		
		// Custom ErrorHandler for the proxy so we can handle unexpected backend drops mid-flight
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] Proxy Error: %v\n", serverUrl.Host, e)
			
			// Mark it as dead immediately so the next request skips it
			// (The health check will bring it back up later if it recovers)
			for _, b := range serverPool.backends {
				if b.URL.String() == serverUrl.String() {
					b.SetAlive(false)
					break
				}
			}
			
			// Try to serve a sensible error to the client
			writer.WriteHeader(http.StatusBadGateway)
			_, _ = writer.Write([]byte("502 Bad Gateway - Backend Server Failed"))
		}

		serverPool.AddBackend(&Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured backend: %s\n", serverUrl)
	}

	// Start the background health checking process
	go serverPool.HealthCheck()

	// Start the load balancer server
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(loadBalance),
	}

	log.Printf("Load Balancer started at port %d\n", port)
	log.Printf("Routing traffic to %d backend servers...\n", len(serverPool.backends))
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
