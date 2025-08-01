import { config } from 'dotenv';
config();

import '@/ai/flows/suggest-resource-context.ts';
import '@/ai/flows/troubleshoot-kubernetes-error.ts';
import '@/ai/flows/summarize-resource-data.ts';