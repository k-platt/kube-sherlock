# Kube Sherlock - Go Backend

This is the Go backend implementation for Kube Sherlock, an AI-powered Kubernetes troubleshooting assistant.

## Features

- **AI-Powered Analysis**: Uses Google Gemini to analyze Kubernetes errors and provide troubleshooting suggestions
- **Resource Gathering**: Automatically gathers relevant Kubernetes resources for context
- **MCP Integration**: Natural language queries with real-time cluster data (e.g., "What is the health of my pods?")
- **CLI Interface**: Command-line tool for direct troubleshooting
- **HTTP API**: REST API endpoints to replace the Firebase frontend functions
- **Multi-Mode**: Both standalone CLI and server mode for frontend integration

## Prerequisites

- Go 1.21 or higher
- Access to a Kubernetes cluster (optional for some features)
- Google AI (Gemini) API key

## Installation

1. Clone the repository and navigate to the backend directory:
```bash
cd backend
```

2. Initialize Go modules and install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o kube-sherlock .
```

## Configuration

### Environment Variables

Set the following environment variables:

```bash
export GEMINI_API_KEY="your-gemini-api-key"
export KUBECONFIG="path/to/your/kubeconfig"  # Optional, defaults to ~/.kube/config
```

### Configuration File

Create a configuration file at `~/.kube-sherlock.yaml`:

```yaml
server:
  host: "localhost"
  port: "8080"

gemini:
  api_key: "your-gemini-api-key"
  model: "gemini-2.0-flash"

kubernetes:
  config_path: "~/.kube/config"
  context: "your-cluster-context"
```

## Usage

### CLI Mode

Analyze a Kubernetes error directly:

```bash
# Basic analysis
./kube-sherlock analyze "ImagePullBackOff"

# With resource gathering
./kube-sherlock analyze --gather-resources --namespace default "CrashLoopBackOff"

# Verbose output
./kube-sherlock analyze --verbose --gather-resources "Pod has unbound immediate PersistentVolumeClaims"
```

### Server Mode

Start the HTTP API server:

```bash
./kube-sherlock server --port 8080 --gemini-api-key "your-api-key"
```

The server will start on `http://localhost:8080` and provide the following endpoints:

- `GET /health` - Health check
- `POST /api/troubleshoot` - Analyze errors (replaces troubleshootKubernetesError)
- `POST /api/suggest-resources` - Get resource suggestions (replaces suggestResourceContext)
- `POST /api/summarize` - Summarize resource data (replaces summarizeResourceData)
- `POST /api/gather-resources` - Gather Kubernetes resources
- `POST /api/query` - **NEW**: Natural language queries with MCP tools

### API Examples

#### Troubleshoot an error:
```bash
curl -X POST http://localhost:8080/api/troubleshoot \
  -H "Content-Type: application/json" \
  -d '{"errorMessage": "ImagePullBackOff"}'
```

#### Suggest resources:
```bash
curl -X POST http://localhost:8080/api/suggest-resources \
  -H "Content-Type: application/json" \
  -d '{"errorDescription": "Pod is failing to start"}'
```

#### Gather resources:
```bash
curl -X POST http://localhost:8080/api/gather-resources \
  -H "Content-Type: application/json" \
  -d '{
    "resourceTypes": ["pods", "deployments", "events"],
    "namespace": "default",
    "labelSelector": "app=myapp"
  }'
```

#### Natural language query (MCP):
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the health of my pods in default namespace?"}'
```

## Frontend Integration

To integrate with the existing Next.js frontend:

1. Update the frontend API calls to point to the Go backend:
   - Replace `troubleshootKubernetesError` calls with `POST /api/troubleshoot`
   - Replace `suggestResourceContext` calls with `POST /api/suggest-resources`
   - Replace `summarizeResourceData` calls with `POST /api/summarize`

2. Update the request/response formats to match the API schemas defined in the handlers.

3. Start the Go backend server before running the frontend.

## Development

### Project Structure

```
backend/
├── main.go                          # Application entry point
├── cmd/                            # CLI commands
│   ├── root.go                     # Root command and configuration
│   ├── server.go                   # HTTP server command
│   └── analyze.go                  # CLI analysis command
├── internal/
│   ├── api/                        # HTTP API handlers
│   │   ├── router.go               # Route configuration
│   │   └── handlers.go             # Request handlers
│   ├── ai/                         # AI service integration
│   │   └── service.go              # Gemini AI client
│   ├── config/                     # Configuration management
│   │   └── config.go               # Config structures and loading
│   └── kubernetes/                 # Kubernetes client
│       └── service.go              # K8s resource operations
├── go.mod                          # Go module definition
└── README.md                       # This file
```

### Adding New Features

1. **New AI Capabilities**: Extend the `ai.Service` with new methods
2. **New Resource Types**: Add support in `kubernetes.Service.GatherResources`
3. **New CLI Commands**: Add new commands in the `cmd/` directory
4. **New API Endpoints**: Add handlers in `api/handlers.go` and routes in `api/router.go`

### Testing

Run tests with:
```bash
go test ./...
```

Build and test the CLI:
```bash
go build -o kube-sherlock .
./kube-sherlock --help
```

Test the server:
```bash
./kube-sherlock server &
curl http://localhost:8080/health
```

## Deployment

### Docker

Create a Dockerfile for containerized deployment:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kube-sherlock .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kube-sherlock .
CMD ["./kube-sherlock", "server"]
```

### Kubernetes

Deploy as a service in your Kubernetes cluster to provide troubleshooting capabilities cluster-wide.

## MCP (Model Context Protocol) Support

Kube Sherlock now supports natural language queries with real-time cluster data:

```bash
# Example queries:
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the health of my pods in default namespace?"}'

curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Show me recent errors in kube-system"}'
```

See `MCP_INTEGRATION.md` for complete documentation and frontend integration examples.

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Use structured logging with zap
5. Handle errors gracefully with proper HTTP status codes

## License

[Add your license information here]
