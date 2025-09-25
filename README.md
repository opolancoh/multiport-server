# Multiport Server

A lightweight HTTP testing server that can listen on multiple ports simultaneously. Perfect for testing network configurations, port availability, and basic HTTP connectivity.

## Features

- 🚀 Listen on multiple ports concurrently
- ✅ Port availability checking with detailed error reporting
- 🌐 External IP detection for easy testing
- 📅 ISO-formatted datetime with timezone in responses
- 🛑 Graceful shutdown with comprehensive resource cleanup
- 📊 Enhanced summary with detailed failure reasons and structured output
- ⚡ Response time tracking in HTTP responses
- 🔒 Permission and port usage detection (root/admin requirements)
- 🧹 Robust error handling with cleanup verification

## Installation

### Option 1: Build locally (requires Go)

```bash
git clone https://github.com/opolancoh/multiport-server.git
cd multiport-server
go build
```

Or run directly:
```bash
go run main.go <port1> <port2> ... <portN>
```

### Option 2: Cross-compile for Raspberry Pi / ARM devices

Build from your development machine and deploy to ARM devices:

```bash
# Build for ARM64 (Raspberry Pi 4, etc.)
GOOS=linux GOARCH=arm64 go build -o multiport-server-arm64

# Copy to your device
scp multiport-server-arm64 pi@your-rpi-ip:/home/pi/

# On your Raspberry Pi/ARM device
chmod +x multiport-server-arm64
./multiport-server-arm64 8080 8081 9000
```

### Option 3: Install Go on Raspberry Pi

If you want to build directly on your Raspberry Pi:

```bash
# Download and install Go
wget https://go.dev/dl/go1.24.3.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.24.3.linux-arm64.tar.gz

# Add to your PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version

# Then build normally
git clone https://github.com/opolancoh/multiport-server.git
cd multiport-server
go build
```

## Usage

Start servers on specific ports:
```bash
./multiport-server 8080 8081 9000
```

Or with `go run`:
```bash
go run main.go 3000 3001 8080 8081
```

## Example Output

```
❌ FAILED PORTS (4)
The following ports could not be used due to issues:

  • 80     → Permission denied (requires root/admin)
  • 443    → Permission denied (requires root/admin)
  • 5100   → Already in use
  • 5101   → Already in use

✅ WORKING PORTS (6)
The following ports are available for use: http://<host>:<port>

  • Localhost       → localhost
  • Local Network   → 192.168.78.200
  • Internet        → 38.188.254.30

Available Ports:
  • 3100
  • 3101
  • 8100
  • 8101
  • 9000
  • 9001

SUMMARY
  • FAILED:  4
  • SUCCESS: 6

Press Ctrl+C to stop...
```

### Graceful Shutdown Output

When terminating the server:
```
Shutting down servers...
🧹 Cleaning up resources...
✅ All servers stopped successfully
```

Or if there are issues during cleanup:
```
🧹 Cleaning up resources...
⚠️  Issues encountered during cleanup:
  • Server on port :8080: shutdown timeout (forced close)
⚠️  Note: Some resources may not have been properly released.
```

## HTTP Response

Each server returns an HTML page showing:
- Server status with running port information
- HTTP response status (200 OK)
- Port number that handled the request
- Current datetime in ISO format with timezone
- Request path
- Response time for performance monitoring

Example response:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Test Server - Port 8080</title>
</head>
<body>
    <h1>HTTP Server Response</h1>
    <p><strong>Status:</strong> Server is currently running at port 8080</p>
    <p><strong>Status:</strong> 200 OK</p>
    <p><strong>Port:</strong> 8080</p>
    <p><strong>DateTime:</strong> 2025-09-24 20:15:30 UTC</p>
    <p><strong>Request Path:</strong> /</p>
    <p><strong>Response Time:</strong> 0.2345 ms</p>
</body>
</html>
```

## Use Cases

- **Network Testing**: Verify multiple ports are accessible with detailed error diagnostics
- **Load Balancer Configuration**: Test port forwarding and routing with clear success/failure reporting
- **Docker/Container Testing**: Check port mapping and connectivity with permission issue detection
- **Development**: Quick HTTP endpoints for testing with response time monitoring
- **CI/CD**: Validate service availability across different ports with structured output for automation
- **System Administration**: Diagnose port conflicts and permission issues quickly
- **Performance Testing**: Monitor response times across different ports

## Error Handling & Troubleshooting

The server provides detailed error messages to help diagnose issues:

### Common Error Messages

- **`Permission denied (requires root/admin)`**: Ports 1-1024 typically require elevated privileges
  ```bash
  sudo ./multiport-server 80 443  # Run with sudo for privileged ports
  ```

- **`Already in use`**: Another service is using the port
  ```bash
  lsof -i :8080  # Check what's using port 8080
  ```

- **`Port out of range`**: Valid port range is 1-65535

- **`Invalid port`**: Non-numeric port arguments

### Cleanup and Resource Management

The server ensures proper cleanup on exit:
- **Graceful shutdown**: Servers stop accepting new connections and finish existing requests
- **Timeout handling**: If servers don't stop within 5 seconds, they're force-closed
- **Error reporting**: Any cleanup issues are reported with specific details
- **Signal handling**: Handles SIGINT, SIGTERM, and SIGABRT properly

## Requirements

- Go 1.21 or higher

## License

MIT License
