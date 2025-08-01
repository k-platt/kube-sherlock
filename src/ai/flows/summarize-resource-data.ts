'use server';
/**
 * @fileOverview Summarizes data gathered from Kubernetes resources using an LLM.
 *
 * - summarizeResourceData - A function that summarizes resource data for diagnosis.
 * - SummarizeResourceDataInput - The input type for the summarizeResourceData function.
 * - SummarizeResourceDataOutput - The return type for the summarizeResourceData function.
 */

import {ai} from '@/ai/genkit';
import {z} from 'genkit';

const SummarizeResourceDataInputSchema = z.object({
  resourceData: z
    .string()
    .describe('Data gathered from Kubernetes resources to be summarized.'),
});
export type SummarizeResourceDataInput = z.infer<typeof SummarizeResourceDataInputSchema>;

const SummarizeResourceDataOutputSchema = z.object({
  summary: z
    .string()
    .describe(
      'A summarized version of the input resource data, highlighting the relevant information for diagnosing issues.'
    ),
});
export type SummarizeResourceDataOutput = z.infer<typeof SummarizeResourceDataOutputSchema>;

export async function summarizeResourceData(
  input: SummarizeResourceDataInput
): Promise<SummarizeResourceDataOutput> {
  return summarizeResourceDataFlow(input);
}

const prompt = ai.definePrompt({
  name: 'summarizeResourceDataPrompt',
  input: {schema: SummarizeResourceDataInputSchema},
  output: {schema: SummarizeResourceDataOutputSchema},
  prompt: `You are an expert Kubernetes troubleshooter. Your task is to summarize the provided data from Kubernetes resources, highlighting only the relevant information for diagnosing issues. Ignore any irrelevant details.

Resource Data:
{{{resourceData}}}`,
});

const summarizeResourceDataFlow = ai.defineFlow(
  {
    name: 'summarizeResourceDataFlow',
    inputSchema: SummarizeResourceDataInputSchema,
    outputSchema: SummarizeResourceDataOutputSchema,
  },
  async input => {
    const {output} = await prompt(input);
    return output!;
  }
);
