// MCP natural language query function

import { kubeSherlockAPI, type MCPQueryResponse } from '@/lib/api-client';

export interface MCPQueryInput {
  query: string;
}

export async function mcpQuery(input: MCPQueryInput): Promise<MCPQueryResponse> {
  return kubeSherlockAPI.mcpQuery(input.query);
}
