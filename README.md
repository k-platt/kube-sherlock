# Kube Sherlock 🕵️‍♂️

A 100% AI generated, AI-powered, Kubernetes troubleshooting assistant that combines natural language queries with real-time cluster analysis. Kube Sherlock features both a powerful Go backend with CLI capabilities and a modern Next.js frontend interface.

## Features

### **Intelligent Analysis**
- **Error Troubleshooting**: Analyze Kubernetes errors and get actionable solutions
- **Natural Language Queries**: Ask questions like "What is the health of my pods in kube-system?"
- **Resource Suggestions**: Get recommendations for gathering relevant troubleshooting context
- **Live Cluster Integration**: Real-time data gathering from your Kubernetes cluster
- **Rich Markdown Responses**: Beautiful formatted output with syntax highlighting

### **Multi-Interface Support**
- **Web Interface**: Modern React-based UI with dual query modes and markdown rendering
- **CLI Tool**: Command-line interface for terminal-based workflows
- **REST API**: HTTP endpoints for integration with other tools
- **MCP Protocol**: Model Context Protocol support for AI tool integration
- **Responsive Design**: Works great on desktop and mobile devices

### **AI-Powered Tools**
- **Pod Health Monitoring**: Check pod status, readiness, and resource usage
- **Deployment Analysis**: Monitor deployment health and replica status  
- **Service Connectivity**: Verify service endpoints and networking
- **Event Monitoring**: Track recent cluster events and issues
- **Log Analysis**: Retrieve and analyze pod logs for troubleshooting

## Quick Start

### Prerequisites
- Go 1.21+ (for backend)
- Node.js 18+ (for frontend)
- Kubernetes cluster access
- Google Gemini API key

### 1. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Set up configuration
cp .env.example .env
# Edit .env with your Gemini API key

# Build the application
go build

# Start the server
./kube-sherlock server --port 8080
```

### 2. Frontend Setup

```bash
# Install dependencies
npm install

# Set up environment
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# Start development server
npm run dev
```

### 3. Usage

#### Web Interface
1. Open http://localhost:3000
2. Choose between two modes:
   - **Troubleshoot Errors**: Traditional error analysis with structured output
   - **Natural Language Query**: Ask questions about your cluster in plain English

#### CLI Usage
```bash
# Analyze a specific error
./kube-sherlock analyze "ImagePullBackOff: Failed to pull image"

# Natural language query
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What pods are failing in my cluster?"}'
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Next.js UI    │    │   Go Backend    │    │  Kubernetes     │
│                 │◄──►│                 │◄──►│  Cluster        │
│ • Troubleshoot  │    │ • REST API      │    │                 │
│ • Natural Lang  │    │ • CLI Tool      │    │ • Live Data     │
│ • Dual Modes    │    │ • MCP Protocol  │    │ • Real Resources│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Google        │
                       │   Gemini AI     │
                       │                 │
                       │ • Analysis      │
                       │ • Tool Selection│
                       │ • Explanations  │
                       └─────────────────┘
```

## 📊 Query Examples

### Natural Language Queries
- `"What is the health of my pods in kube-system namespace?"`
- `"Show me recent events in the default namespace"`
- `"Which deployments are not ready?"`
- `"Get logs for failing pods"`
- `"What services are available and their endpoints?"`

### Traditional Error Analysis
- `"ImagePullBackOff: Failed to pull image 'nginx:latest'"`
- `"CrashLoopBackOff: Container failed to start"`
- `"Service Unavailable: Cannot connect to backend"`

## 🛠️ Development

### Backend Development
```bash
cd backend

# Run tests
go test ./...

# Build for production
go build -o kube-sherlock

# Run with verbose logging
./kube-sherlock server --verbose
```

### Frontend Development
```bash
# Start development server
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check
```

## 📋 API Documentation

### Endpoints

#### `POST /api/troubleshoot`
Analyze Kubernetes errors and get structured troubleshooting guidance.

**Request:**
```json
{
  "error_message": "ImagePullBackOff: Failed to pull image"
}
```

**Response:**
```json
{
  "potentialCauses": ["Image not found", "Registry authentication failed"],
  "suggestedSolutions": ["Check image name", "Verify registry credentials"]
}
```

#### `POST /api/query` (MCP-enabled)
Process natural language queries with live cluster data integration.

**Request:**
```json
{
  "query": "What is the health of my pods in kube-system namespace?"
}
```

**Response:**
```json
{
  "response": "Your kube-system namespace has 8 pods running...",
  "usedTool": true,
  "toolUsed": "get_pod_health",
  "rawData": "detailed cluster data..."
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Powered by [Google Gemini AI](https://ai.google.dev/)
- Built with [Go](https://golang.org/), [Next.js](https://nextjs.org/), and [Kubernetes](https://kubernetes.io/)
- UI components from [shadcn/ui](https://ui.shadcn.com/)
- CLI framework: [Cobra](https://github.com/spf13/cobra)

---

**Ready to troubleshoot your Kubernetes cluster with AI? Get started now!** 🚀
