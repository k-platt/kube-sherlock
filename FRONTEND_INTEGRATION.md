# Frontend Integration Guide

This guide explains how to integrate the Go backend with the existing Next.js frontend.

## Quick Start

1. **Start the Go Backend Server:**
```bash
cd backend
export GEMINI_API_KEY="your-gemini-api-key"
./kube-sherlock server --port 8080
```

2. **Update Frontend API Calls:**

The Go backend provides REST endpoints that replace the Firebase Genkit functions:

| Original Function | New Endpoint | Method |
|-------------------|--------------|--------|
| `troubleshootKubernetesError` | `/api/troubleshoot` | POST |
| `suggestResourceContext` | `/api/suggest-resources` | POST |
| `summarizeResourceData` | `/api/summarize` | POST |

## Code Changes Required

### 1. Update API Base URL

Create or update your API configuration to point to the Go backend:

```typescript
// src/lib/api.ts
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = {
  troubleshoot: (errorMessage: string) => 
    fetch(`${API_BASE_URL}/api/troubleshoot`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ errorMessage }),
    }),
  
  suggestResources: (errorDescription: string) =>
    fetch(`${API_BASE_URL}/api/suggest-resources`, {
      method: 'POST', 
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ errorDescription }),
    }),
    
  summarize: (resourceData: string) =>
    fetch(`${API_BASE_URL}/api/summarize`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ resourceData }),
    }),
};
```

### 2. Update KubeSherlock Component

Replace the AI flow imports with API calls:

```typescript
// src/components/kube-sherlock.tsx

// Remove these imports:
// import { troubleshootKubernetesError, type TroubleshootKubernetesErrorOutput } from "@/ai/flows/troubleshoot-kubernetes-error";
// import { suggestResourceContext, type SuggestResourceContextOutput } from "@/ai/flows/suggest-resource-context";

// Add this import:
import { api } from "@/lib/api";

// Update the onSubmit function:
function onSubmit(values: z.infer<typeof formSchema>) {
  setAnalysisResult(null);
  setSuggestedResources(null);

  startTransition(async () => {
    try {
      if (verbosity) {
        console.log("Starting investigation with error: ", values.errorMessage);
      }
      
      // Call the Go backend instead of Genkit functions
      const [analysisResponse, suggestionsResponse] = await Promise.all([
        api.troubleshoot(values.errorMessage),
        api.suggestResources(values.errorMessage),
      ]);

      if (!analysisResponse.ok || !suggestionsResponse.ok) {
        throw new Error("Failed to get a response from the AI assistant.");
      }

      const analysis = await analysisResponse.json();
      const suggestions = await suggestionsResponse.json();
      
      setAnalysisResult(analysis);
      setSuggestedResources(suggestions);

    } catch (error) {
      console.error("Error during analysis:", error);
      toast({
        variant: "destructive",
        title: "Analysis Failed",
        description: error instanceof Error ? error.message : "An unknown error occurred.",
      });
    }
  });
}
```

### 3. Update Type Definitions

The response types are compatible, but you may want to define them explicitly:

```typescript
// src/types/api.ts
export interface TroubleshootResponse {
  potentialCauses: string[];
  suggestedSolutions: string[];
}

export interface SuggestResourcesResponse {
  suggestedResources: string[];
  reasoning: string;
}

export interface SummarizeResponse {
  summary: string;
}
```

### 4. Environment Variables

Add to your `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

For production, update this to your deployed Go backend URL.

## Development Workflow

1. **Start Backend:**
```bash
cd backend
export GEMINI_API_KEY="your-key"
./kube-sherlock server
```

2. **Start Frontend:**
```bash
cd .. # back to project root
npm run dev
```

3. **Test Integration:**
- Open http://localhost:9002 (or your Next.js port)
- Enter a Kubernetes error message
- Verify the analysis works

## Additional Features

The Go backend provides additional capabilities not available in the original frontend:

### Resource Gathering

You can now gather actual Kubernetes resources:

```typescript
const gatherResources = async (resourceTypes: string[], namespace: string) => {
  const response = await fetch(`${API_BASE_URL}/api/gather-resources`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      resourceTypes,
      namespace,
      labelSelector: '', // optional
    }),
  });
  return response.json();
};
```

### Health Check

Monitor backend health:

```typescript
const healthCheck = () => fetch(`${API_BASE_URL}/health`);
```

## Production Deployment

1. **Backend:** Deploy the Go binary with environment variables
2. **Frontend:** Update `NEXT_PUBLIC_API_URL` to your production backend URL
3. **CORS:** The backend includes CORS headers for cross-origin requests

## Troubleshooting

- **CORS Issues:** Ensure the backend is running and CORS is properly configured
- **API Key:** Verify `GEMINI_API_KEY` is set for the backend
- **Network:** Check that the frontend can reach the backend URL
- **Logs:** Use `--verbose` flag for detailed backend logging

## CLI Usage

The backend also provides a CLI for direct troubleshooting:

```bash
# Direct analysis
./kube-sherlock analyze "ImagePullBackOff" --gemini-api-key "your-key"

# With resource gathering
./kube-sherlock analyze "CrashLoopBackOff" --gather-resources --namespace default --gemini-api-key "your-key"
```
