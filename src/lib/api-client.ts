// API client for kube-sherlock Go backend

export interface TroubleshootResponse {
  potentialCauses: string[];
  suggestedSolutions: string[];
}

export interface SuggestResourcesResponse {
  suggestedResources: string[];
  reasoning: string;
}

export interface MCPQueryResponse {
  response: string;
  usedTool: boolean;
  toolUsed?: string;
  rawData?: string;
  error?: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class KubeSherlockAPI {
  private async fetchAPI<T>(endpoint: string, data: any): Promise<T> {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API request failed: ${response.status} ${errorText}`);
    }

    return response.json();
  }

  async troubleshoot(errorMessage: string): Promise<TroubleshootResponse> {
    return this.fetchAPI<TroubleshootResponse>('/api/troubleshoot', {
      error_message: errorMessage,
    });
  }

  async suggestResources(errorDescription: string): Promise<SuggestResourcesResponse> {
    return this.fetchAPI<SuggestResourcesResponse>('/api/suggest-resources', {
      error_description: errorDescription,
    });
  }

  async mcpQuery(query: string): Promise<MCPQueryResponse> {
    return this.fetchAPI<MCPQueryResponse>('/api/query', {
      query: query,
    });
  }
}

export const kubeSherlockAPI = new KubeSherlockAPI();
