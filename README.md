# Golance

A lightweight HTTP/HTTPS load balancer written in Go. Golance distributes incoming requests across multiple backend servers with support for session persistence via cookies.

## Features

- **HTTP and HTTPS Support**: Listens on both HTTP (port 8080) and HTTPS (port 8443)
- **Multiple Backend Support**: Distributes requests across configurable backend servers
- **Session Persistence**: Uses cookie-based session affinity (`LB_GOLANCE` cookie) to maintain client-backend relationships
- **Random Load Balancing**: Randomly selects a backend when no session cookie is present
- **Mixed Protocol Backends**: Supports both HTTP and HTTPS backend servers
- **Request Forwarding**: Properly forwards HTTP requests with header manipulation

## Architecture

Golance acts as a reverse proxy that:
1. Accepts incoming HTTP/HTTPS connections
2. Parses HTTP requests
3. Selects a backend server (using cookie persistence or random selection)
4. Forwards the request to the selected backend
5. Returns the backend's response to the client

## Requirements

- Go 1.25.4 or later
- TLS certificates (`cert.pem` and `key.pem`) for HTTPS support

## Installation

1. Clone the repository:
```bash
git clone https://github.com/rurueuh/Golance
cd Golance
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
go build -o golance
```

## Configuration

Edit the `backends` slice in `golance.go` to configure your backend servers:

```go
var backends = []Backend{
    {Address: "rurueuh.fr:443", IsHTTPS: true},
    {Address: "example.com:443", IsHTTPS: true},
    {Address: "example.com:80", IsHTTPS: false},
}
```

Each backend requires:
- `Address`: The backend server address (host:port format)
- `IsHTTPS`: Boolean indicating if the backend uses HTTPS

## Usage

1. **Prepare TLS certificates** (required for HTTPS):
   - Place `cert.pem` and `key.pem` in the project root directory
   - These certificates are used for the HTTPS listener on port 8443

2. **Run the load balancer**:
```bash
./golance
```

Or on Windows:
```bash
golance.exe
```

The server will start listening on:
- HTTP: `:8080`
- HTTPS: `:8443`

3. **Test the load balancer**:
   - curl:
   ```bash
   curl http://localhost:8080
   curl -k https://localhost:8443
   ```

## How It Works

### Request Flow

1. Client sends request to Golance (HTTP or HTTPS)
2. Golance parses the HTTP request headers
3. Checks for `LB_GOLANCE` cookie to determine backend selection
4. If cookie exists and is valid, routes to that backend
5. If no cookie or invalid cookie, randomly selects a backend
6. Forwards request to selected backend (HTTP or HTTPS as configured)
7. Returns backend response to client

### Session Persistence

Golance uses the `LB_GOLANCE` cookie to maintain session affinity. The cookie value contains the index of the backend server. When a client includes this cookie, requests are routed to the same backend server.

## Project Structure

```
Golance/
├── golance.go      # Main entry point, server setup, backend configuration
├── request.go      # HTTP request parsing and connection handling
├── response.go     # Response forwarding to backends
├── header.go       # HTTP header parsing and formatting
├── go/
│   └── get.go      # Test client for load balancer
├── go.mod          # Go module dependencies
└── go.sum          # Dependency checksums
```

## Supported HTTP Methods

Golance validates and supports the following HTTP methods:
- OPTIONS
- GET
- HEAD
- POST
- PUT
- DELETE
- TRACE
- CONNECT

## Security Notes

- The HTTPS backend connections use `InsecureSkipVerify: true`, which means SSL certificate verification is disabled. **Use with caution in production environments.**
- Ensure your TLS certificates (`cert.pem` and `key.pem`) are properly secured and not committed to version control.

## Development

### Dependencies

- `golang.org/x/text`: Used for proper HTTP header formatting (Title case conversion)

### Building

```bash
go build -o golance
```

