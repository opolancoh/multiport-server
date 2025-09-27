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

### Option 2: Cross-compile for different platforms

Build for various operating systems and architectures:

```bash
# Linux ARM64 (Raspberry Pi 4, etc.)
GOOS=linux GOARCH=arm64 go build -o multiport-server-linux-arm64

# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o multiport-server-linux-amd64

# macOS ARM64 (Apple Silicon M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o multiport-server-darwin-arm64

# macOS Intel (older Macs)
GOOS=darwin GOARCH=amd64 go build -o multiport-server-darwin-amd64

# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o multiport-server-windows-amd64.exe

# Windows 32-bit
GOOS=windows GOARCH=386 go build -o multiport-server-windows-386.exe
```

#### macOS Usage
```bash
# Make executable and run
chmod +x multiport-server-darwin-arm64
./multiport-server-darwin-arm64 8080 8081 9000
```

#### Windows Usage
```cmd
# Run on Windows (Command Prompt or PowerShell)
multiport-server-windows-amd64.exe 8080 8081 9000
```

### Option 3: Cross-compile for Raspberry Pi / ARM devices

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

## Background Operation & Monitoring

For running the server in background (especially useful when accessing via SSH):

### Start in Background

```bash
# Start server in background with nohup
nohup ./multiport-server-arm64 8080 8081 9000 &
```

This command:
- `nohup` - Keeps the process running even if you disconnect from SSH
- `&` - Runs the process in background, giving you back the console
- Output is automatically redirected to `nohup.out` file

### Monitor Server Output

```bash
# View all output from the server
cat nohup.out

# View output in real-time (follow mode)
tail -f nohup.out
```

### Check Running Process

```bash
# Find your server process
ps aux | grep multiport-server
```

Example output:
```
pheidias    2555  0.0  0.1 1230088 6656 pts/0    Sl   21:40   0:00 ./multiport-server-arm64 8080 8081 9000
```

This shows:
- Process ID: `2555`
- Memory usage: `6656 KB`
- Status: `Sl` (Sleeping, multi-threaded - waiting for connections)
- Start time: `21:40`

### Check Listening Ports

```bash
# View all listening TCP ports
netstat -tln

# Filter for your specific ports
netstat -tln | grep -E '8080|8081|9000'
```

Example output:
```
tcp6       0      0 :::8080                :::*                    LISTEN
tcp6       0      0 :::8081                :::*                    LISTEN
tcp6       0      0 :::9000                :::*                    LISTEN
```

### Stop Server Gracefully

```bash
# Send interrupt signal (same as Ctrl+C)
kill -INT 2555

# Or use SIGTERM for standard termination
kill -TERM 2555

# Or simply (defaults to SIGTERM)
kill 2555
```

Replace `2555` with your actual process ID from `ps aux | grep multiport-server`.

The server will:
1. Print "Shutting down servers..."
2. Stop accepting new connections
3. Finish processing existing requests
4. Clean up all resources
5. Release all ports

### Verify Shutdown

```bash
# Check if process is still running
ps aux | grep multiport-server

# Verify ports are released
netstat -tln | grep -E '8080|8081|9000'
```

Both commands should return no results if the server stopped properly.

## Requirements

- Go 1.21 or higher

## License

MIT License
