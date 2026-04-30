# Go-LB (Go Load Balancer)

A high-performance, concurrent Layer 7 Reverse Proxy and Load Balancer built purely in Go, without external dependencies.

This project was built to demonstrate advanced idiomatic Go concepts including concurrency, thread-safe memory management, and network programming.

## Features

- **Round-Robin Load Balancing:** Evenly distributes incoming HTTP requests across a pool of backend servers.
- **Active Health Checking:** Uses background Goroutines to continually monitor backend health. If a server goes down, it is automatically removed from the active routing pool.
- **Fault Tolerance:** Thread-safe state management (`sync.RWMutex` and `sync/atomic`) ensures the proxy safely handles backend failures and recoveries mid-flight without race conditions.
- **Graceful Failover:** Proxy errors instantly mark a server as dead, skipping the bad node until the active health checker verifies it has recovered.

## Architecture Highlights

- **`net/http/httputil`**: Utilizes Go's built-in ReverseProxy for robust request forwarding.
- **Goroutines & WaitGroups**: The active health checker spins up concurrent Go routines for every registered backend, using `sync.WaitGroup` to synchronize the checks.
- **Atomic Operations**: Uses `sync/atomic` for incrementing the routing index, guaranteeing safe, lock-free Round-Robin execution under heavy load.

## Getting Started

### 1. Run the Backend Simulators

To test the load balancer, spin up a few instances of the simulated microservices on different ports. Open multiple terminal tabs and run:

```bash
# Terminal 1
go run cmd/sim/main.go --port=8081 --name="Backend-1"

# Terminal 2
go run cmd/sim/main.go --port=8082 --name="Backend-2"

# Terminal 3
go run cmd/sim/main.go --port=8083 --name="Backend-3"
```

### 2. Start the Load Balancer

In a new terminal tab, start the `go-lb` server and point it to the backends you just spun up:

```bash
go run *.go --port=8080 --backends=http://localhost:8081,http://localhost:8082,http://localhost:8083
```

### 3. Test the Routing

Send requests to the load balancer at `http://localhost:8080`. You will see the responses alternating between the 3 backends!

```bash
curl http://localhost:8080
```

### 4. Test Fault Tolerance (The cool part)

1. Kill one of the backend simulator processes (e.g., `Backend-2` on port 8082).
2. Look at the Load Balancer logs. You will see the active health checker immediately detect the failure: `Health Check: [http://localhost:8082] is now down`.
3. Send more `curl` requests. The load balancer will automatically skip the dead server and only route traffic between `Backend-1` and `Backend-3`.
4. Restart the `Backend-2` process. The health checker will detect the recovery, and traffic will seamlessly resume routing to it!

## Author
[Owais Noe](https://github.com/noekunn)
