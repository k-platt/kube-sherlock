# Model Context Protocol (MCP) Integration

This document explains the MCP integration that allows natural language queries with real-time Kubernetes cluster data.

## Overview

The MCP integration enables queries like:
- "What is the health of my pods in default namespace?"
- "Show me the deployment status for my production applications"
- "Are there any recent errors in the kube-system namespace?"
- "Get logs from the failing pod in my-app namespace"

## How It Works

1. **Natural Language Processing**: User submits a query in plain English
2. **Tool Selection**: AI analyzes the query and selects appropriate Kubernetes tools
3. **Data Gathering**: Real-time cluster data is collected using Kubernetes APIs
4. **AI Analysis**: Gathered data is analyzed by AI to provide insights
5. **Response**: User receives a comprehensive answer with cluster context

## Available MCP Tools

### get_pod_health
- **Purpose**: Get health status of pods in a namespace
- **Parameters**: 
  - `namespace` (optional): Target namespace (default: "default")
  - `labelSelector` (optional): Filter pods by labels

### get_deployment_status
- **Purpose**: Get deployment status and replica information
- **Parameters**:
  - `namespace` (optional): Target namespace (default: "default") 
  - `deploymentName` (optional): Specific deployment name

### get_service_endpoints
- **Purpose**: Get service endpoints and connectivity status
- **Parameters**:
  - `namespace` (optional): Target namespace (default: "default")
  - `serviceName` (optional): Specific service name

### get_recent_events
- **Purpose**: Get recent Kubernetes events
- **Parameters**:
  - `namespace` (optional): Target namespace (default: "default")
  - `resourceName` (optional): Filter events for specific resource

### get_pod_logs
- **Purpose**: Get logs from a specific pod
- **Parameters**:
  - `namespace` (optional): Target namespace (default: "default")
  - `podName` (required): Pod name to get logs from
  - `containerName` (optional): Specific container name
  - `lines` (optional): Number of lines to retrieve (default: 100)

## API Usage

### Endpoint
```
POST /api/query
```

### Request Format
```json
{
  "query": "What is the health of my pods in default namespace?"
}
```

### Response Format
```json
{
  "response": "Based on the current cluster data, you have 3 pods in the default namespace. All pods are running healthy with the following status...",
  "usedTool": true,
  "toolUsed": "get_pod_health",
  "rawData": "...", // Optional: raw cluster data
  "error": ""       // Optional: error message if any
}
```

## Example Queries

### Pod Health Check
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the health of my pods in default namespace?"}'
```

### Deployment Status
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Show me the status of all deployments in production namespace"}'
```

### Recent Events
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Are there any recent errors or warnings in kube-system?"}'
```

### Pod Logs
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "Get the last 50 lines of logs from pod myapp-123 in production namespace"}'
```

## Frontend Integration

To add MCP support to the frontend:

### 1. Add MCP Query Function
```typescript
// src/lib/api.ts
export const api = {
  // ... existing functions
  
  mcpQuery: (query: string) =>
    fetch(`${API_BASE_URL}/api/query`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query }),
    }),
};
```

### 2. Add MCP Query Component
```typescript
// src/components/mcp-query.tsx
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";

export function MCPQuery() {
  const [query, setQuery] = useState("");
  const [response, setResponse] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!query.trim()) return;
    
    setLoading(true);
    try {
      const res = await api.mcpQuery(query);
      const data = await res.json();
      setResponse(data);
    } catch (error) {
      console.error("MCP query failed:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ask About Your Cluster</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <Textarea
          placeholder="What is the health of my pods in default namespace?"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <Button onClick={handleSubmit} disabled={loading}>
          {loading ? "Analyzing..." : "Ask"}
        </Button>
        
        {response && (
          <div className="mt-4 p-4 bg-muted rounded-md">
            <p className="whitespace-pre-wrap">{response.response}</p>
            {response.usedTool && (
              <p className="text-sm text-muted-foreground mt-2">
                Used tool: {response.toolUsed}
              </p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
```

## Configuration

No additional configuration is required. MCP tools are automatically available when:
1. Kubernetes service is successfully initialized
2. Cluster connectivity is established
3. AI service (Gemini) is configured

## Benefits

### Real-Time Data
- Live cluster information, not static responses
- Current pod status, deployments, and events
- Fresh log data

### Natural Language Interface
- No need to remember kubectl commands
- Conversational troubleshooting
- Context-aware responses

### Intelligent Analysis
- AI interprets cluster data
- Provides insights and recommendations
- Correlates multiple data sources

## Technical Implementation

### Architecture
```
User Query → AI Analysis → Tool Selection → K8s API → Data Gathering → AI Analysis → Response
```

### Tools Framework
- Extensible tool system
- JSON schema validation
- Error handling and fallbacks
- Structured responses

### Integration Points
- Kubernetes client-go library
- Google Gemini AI API
- MCP protocol standards
- REST API endpoints

## Troubleshooting

### Common Issues

**"Kubernetes service not available"**
- Ensure kubectl connectivity to cluster
- Verify kubeconfig is valid
- Check cluster permissions

**"MCP service not available"**
- Kubernetes service must be initialized first
- Check server logs for initialization errors

**"Tool execution failed"**
- Verify namespace exists
- Check RBAC permissions
- Review resource names

### Debugging
- Use `--verbose` flag for detailed logs
- Check `/health` endpoint
- Verify API responses with curl

## Future Enhancements

Potential additions:
- More Kubernetes resource types
- Custom tool definitions
- Resource modification capabilities
- Multi-cluster support
- Persistent conversation context
