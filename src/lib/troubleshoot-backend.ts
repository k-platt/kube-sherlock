// Drop-in replacement for Firebase Genkit troubleshoot function

import { kubeSherlockAPI } from '@/lib/api-client';

export interface TroubleshootKubernetesErrorInput {
  errorMessage: string;
}

export interface TroubleshootKubernetesErrorOutput {
  potentialCauses: string[];
  suggestedSolutions: string[];
}

export async function troubleshootKubernetesError(
  input: TroubleshootKubernetesErrorInput
): Promise<TroubleshootKubernetesErrorOutput> {
  return kubeSherlockAPI.troubleshoot(input.errorMessage);
}
