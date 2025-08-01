'use server';

/**
 * @fileOverview Suggests Kubernetes resources that would provide helpful context for troubleshooting.
 *
 * - suggestResourceContext - A function that suggests Kubernetes resources for troubleshooting.
 * - SuggestResourceContextInput - The input type for the suggestResourceContext function.
 * - SuggestResourceContextOutput - The return type for the suggestResourceContext function.
 */

import {ai} from '@/ai/genkit';
import {z} from 'genkit';

const SuggestResourceContextInputSchema = z.object({
  errorDescription: z
    .string()
    .describe('Description of the error encountered in the Kubernetes cluster.'),
});
export type SuggestResourceContextInput = z.infer<typeof SuggestResourceContextInputSchema>;

const SuggestResourceContextOutputSchema = z.object({
  suggestedResources: z
    .array(z.string())
    .describe(
      'List of Kubernetes resources (e.g., pod logs, deployment configurations) that could provide helpful context for troubleshooting.'
    ),
  reasoning: z
    .string()
    .describe('Explanation of why the suggested resources are relevant.'),
});
export type SuggestResourceContextOutput = z.infer<typeof SuggestResourceContextOutputSchema>;

export async function suggestResourceContext(
  input: SuggestResourceContextInput
): Promise<SuggestResourceContextOutput> {
  return suggestResourceContextFlow(input);
}

const prompt = ai.definePrompt({
  name: 'suggestResourceContextPrompt',
  input: {schema: SuggestResourceContextInputSchema},
  output: {schema: SuggestResourceContextOutputSchema},
  prompt: `You are a Kubernetes troubleshooting expert. Given the following error description, suggest which Kubernetes resources (e.g., logs from specific pods, deployment configurations) would provide helpful context for troubleshooting.

Error Description: {{{errorDescription}}}

Respond with a list of suggested resources and a brief explanation of why each resource is relevant. Focus on resources that would help diagnose the root cause of the error.

Example Output:
{
  "suggestedResources": ["pod/example-pod logs", "deployment/example-deployment configuration", "service/example-service description"],
  "reasoning": "Pod logs may contain error messages. Deployment configuration can show misconfigurations. Service description can show service unavailable issues."
}
`,
});

const suggestResourceContextFlow = ai.defineFlow(
  {
    name: 'suggestResourceContextFlow',
    inputSchema: SuggestResourceContextInputSchema,
    outputSchema: SuggestResourceContextOutputSchema,
  },
  async input => {
    const {output} = await prompt(input);
    return output!;
  }
);
