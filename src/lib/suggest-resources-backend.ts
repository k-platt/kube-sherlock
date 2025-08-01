// Drop-in replacement for Firebase Genkit suggest-resource-context function

import { kubeSherlockAPI } from '@/lib/api-client';

export interface SuggestResourceContextInput {
  errorDescription: string;
}

export interface SuggestResourceContextOutput {
  suggestedResources: string[];
  reasoning: string;
}

export async function suggestResourceContext(
  input: SuggestResourceContextInput
): Promise<SuggestResourceContextOutput> {
  return kubeSherlockAPI.suggestResources(input.errorDescription);
}
