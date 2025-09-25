package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run server.go <port1> <port2> ... <portN>")
		os.Exit(1)
	}

	ports := []int{}
	failedPorts := []PortFailure{}

	// Parse and validate ports
	for _, arg := range os.Args[1:] {
		port, err := strconv.Atoi(arg)
		if err != nil {
			failedPorts = append(failedPorts, PortFailure{Port: 0, Reason: fmt.Sprintf("Invalid port: %s", arg)})
			continue
		}
		if port < 1 || port > 65535 {
			failedPorts = append(failedPorts, PortFailure{Port: port, Reason: "Port out of range"})
			continue
		}

		// Check if port is available
		if err := checkPortAvailability(port); err != nil {
			failedPorts = append(failedPorts, PortFailure{Port: port, Reason: fmt.Sprintf("Port unavailable: %v", err)})
		} else {
			ports = append(ports, port)
		}
	}

	// Get IPs for testing examples
	localIP := getLocalIP()
	externalIP := getExternalIP()

	// Start servers
	var wg sync.WaitGroup
	var mu sync.Mutex
	servers := make([]*http.Server, 0, len(ports))
	successfulPorts := []int{}

	for _, port := range ports {
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: createHandler(port),
		}
		servers = append(servers, server)

		wg.Add(1)
		go func(s *http.Server, p int) {
			defer wg.Done()
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				// Server failed to start after initial check
				mu.Lock()
				failedPorts = append(failedPorts, PortFailure{Port: p, Reason: fmt.Sprintf("Server failed to start: %v", err)})
				mu.Unlock()
			}
		}(server, port)

		// Give server a moment to start
		time.Sleep(10 * time.Millisecond)
		if isServerRunning(port) {
			successfulPorts = append(successfulPorts, port)
		} else {
			// Server didn't start successfully, add to failed ports if not already there
			mu.Lock()
			// Check if this port is already in failedPorts
			found := false
			for _, fp := range failedPorts {
				if fp.Port == port {
					found = true
					break
				}
			}
			if !found {
				failedPorts = append(failedPorts, PortFailure{Port: port, Reason: "Server failed to start or is not responding"})
			}
			mu.Unlock()
		}
	}

	// Wait for all server goroutines to be ready
	time.Sleep(50 * time.Millisecond)

	// Print summary
	printSummary(failedPorts, successfulPorts, localIP, externalIP)

	if len(successfulPorts) == 0 {
		fmt.Println("⏳ Waiting for server goroutines to complete...")

		// Wait for goroutines with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// All goroutines completed normally
		case <-time.After(2 * time.Second):
			// Timeout waiting for goroutines
			fmt.Println("⚠️  Timeout waiting for server goroutines, proceeding with cleanup...")
		}

		cleanupResources(servers)
		os.Exit(1)
	}

	// Setup cleanup on program termination
	setupCleanupOnExit(servers)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nPress Ctrl+C to stop...")
	<-sigChan
	fmt.Println("\nShutting down servers...")

	// Shutdown all servers gracefully
	cleanupResources(servers)
}

type PortFailure struct {
	Port   int
	Reason string
}

func setupCleanupOnExit(servers []*http.Server) {
	// Ensure cleanup happens even if program exits unexpectedly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	go func() {
		<-c
		fmt.Println("\n🚨 Unexpected termination detected, cleaning up...")
		cleanupResources(servers)
		os.Exit(1)
	}()
}

func cleanupResources(servers []*http.Server) {
	if len(servers) == 0 {
		fmt.Println("✅ No servers to clean up")
		return
	}

	fmt.Println("🧹 Cleaning up resources...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	var shutdownWg sync.WaitGroup
	var shutdownErrors []string
	var errorMu sync.Mutex

	for i, server := range servers {
		shutdownWg.Add(1)
		go func(s *http.Server, index int) {
			defer shutdownWg.Done()

			// Attempt graceful shutdown
			if err := s.Shutdown(shutdownCtx); err != nil {
				errorMu.Lock()
				if err == context.DeadlineExceeded {
					shutdownErrors = append(shutdownErrors, fmt.Sprintf("Server on port %s: shutdown timeout (forced close)", s.Addr))
					// Force close if timeout
					if closeErr := s.Close(); closeErr != nil {
						shutdownErrors = append(shutdownErrors, fmt.Sprintf("Server on port %s: force close error: %v", s.Addr, closeErr))
					}
				} else {
					shutdownErrors = append(shutdownErrors, fmt.Sprintf("Server on port %s: shutdown error: %v", s.Addr, err))
				}
				errorMu.Unlock()
			}
		}(server, i)
	}

	shutdownWg.Wait()

	if len(shutdownErrors) > 0 {
		fmt.Println("⚠️  Issues encountered during cleanup:")
		for _, errMsg := range shutdownErrors {
			fmt.Printf("  • %s\n", errMsg)
		}
		fmt.Println("⚠️  Note: Some resources may not have been properly released.")
	} else {
		fmt.Println("✅ All servers stopped successfully")
	}
}

func createHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		datetime := startTime.Format("2006-01-02 15:04:05 MST")

		responseTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Test Server - Port %d</title>
</head>
<body>
    <h1>HTTP Server Response</h1>
    <p><strong>Status:</strong> Server is currently running at port %d</p>
    <p><strong>Status:</strong> 200 OK</p>
	<p><strong>Port:</strong> %d</p>
    <p><strong>DateTime:</strong> %s</p>
    <p><strong>Request Path:</strong> %s</p>
    <p><strong>Response Time:</strong> %.4f ms</p>
</body>
</html>`, port, port, port, datetime, r.URL.Path, responseTime)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}

func checkPortAvailability(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	ln.Close()
	return nil
}

func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}

func isServerRunning(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func getExternalIP() string {
	// Try to get external IP
	resp, err := http.Get("http://ipv4.icanhazip.com")
	if err != nil {
		return "N/A"
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "N/A"
	}

	return strings.TrimSpace(string(ip))
}

func formatFailureReason(reason string) string {
	// Clean up common error messages to be more user-friendly
	if strings.Contains(reason, "permission denied") {
		return "Permission denied (requires root/admin)"
	}
	if strings.Contains(reason, "address already in use") {
		return "Already in use"
	}
	if strings.Contains(reason, "Port unavailable:") {
		// Extract the actual error after "Port unavailable: listen tcp :<port>: bind: "
		parts := strings.Split(reason, "bind: ")
		if len(parts) > 1 {
			errorMsg := parts[1]
			if strings.Contains(errorMsg, "permission denied") {
				return "Permission denied (requires root/admin)"
			}
			if strings.Contains(errorMsg, "address already in use") {
				return "Already in use"
			}
			return errorMsg
		}
	}
	return reason
}

func printSummary(failedPorts []PortFailure, successfulPorts []int, localIP, externalIP string) {
	fmt.Println() // Initial break line

	// Print failed ports
	if len(failedPorts) > 0 {
		fmt.Printf("❌ FAILED PORTS (%d)\n", len(failedPorts))
		fmt.Println("The following ports could not be used due to issues:")
		fmt.Println()
		for _, failure := range failedPorts {
			if failure.Port == 0 {
				fmt.Printf("  • %s\n", failure.Reason)
			} else {
				reason := formatFailureReason(failure.Reason)
				fmt.Printf("  • %-6d → %s\n", failure.Port, reason)
			}
		}
		fmt.Println()
	}

	// Print successful ports
	if len(successfulPorts) > 0 {
		fmt.Printf("✅ WORKING PORTS (%d)\n", len(successfulPorts))
		fmt.Println("The following ports are available for use: http://<host>:<port>")
		fmt.Println()

		fmt.Printf("  • %-15s → localhost\n", "Localhost")
		fmt.Printf("  • %-15s → %s\n", "Local Network", localIP)
		if externalIP != "N/A" {
			fmt.Printf("  • %-15s → %s\n", "Internet", externalIP)
		} else {
			fmt.Printf("  • %-15s → Unable to determine\n", "Internet")
		}

		fmt.Println()
		fmt.Println("Available Ports:")
		for _, port := range successfulPorts {
			fmt.Printf("  • %d\n", port)
		}
		fmt.Println()
	}

	// Print summary counts
	fmt.Println("SUMMARY")
	fmt.Printf("  • FAILED:  %d\n", len(failedPorts))
	fmt.Printf("  • SUCCESS: %d\n", len(successfulPorts))
	fmt.Println()
}
