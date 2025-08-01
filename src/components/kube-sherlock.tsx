"use client";

import { useState, useTransition } from "react";
// Use Go backend instead of Firebase Genkit
import { troubleshootKubernetesError, type TroubleshootKubernetesErrorOutput } from "@/lib/troubleshoot-backend";
import { suggestResourceContext, type SuggestResourceContextOutput } from "@/lib/suggest-resources-backend";
import { mcpQuery, type MCPQueryInput } from "@/lib/mcp-query";
import { type MCPQueryResponse } from "@/lib/api-client";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { BrainCircuit, FileText, Lightbulb, Loader2, Sparkles, Wand2, Wrench } from "lucide-react";
import { Skeleton } from "./ui/skeleton";

const formSchema = z.object({
  errorMessage: z.string().min(10, {
    message: "Please enter a more detailed message (at least 10 characters).",
  }),
});

export function KubeSherlock() {
  const [isPending, startTransition] = useTransition();
  const [verbosity, setVerbosity] = useState(false);
  const [queryMode, setQueryMode] = useState<'troubleshoot' | 'natural'>('troubleshoot');
  const [analysisResult, setAnalysisResult] = useState<TroubleshootKubernetesErrorOutput | null>(null);
  const [suggestedResources, setSuggestedResources] = useState<SuggestResourceContextOutput | null>(null);
  const [mcpResult, setMcpResult] = useState<MCPQueryResponse | null>(null);
  const { toast } = useToast();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      errorMessage: "",
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    // Clear previous results
    setAnalysisResult(null);
    setSuggestedResources(null);
    setMcpResult(null);

    startTransition(async () => {
      try {
        if (verbosity) {
          console.log(`Starting ${queryMode} with message: `, values.errorMessage);
        }

        if (queryMode === 'natural') {
          // Use MCP natural language query
          const mcpResponse = await mcpQuery({ query: values.errorMessage });
          
          if (!mcpResponse) {
            throw new Error("Failed to get a response from the AI assistant.");
          }
          
          setMcpResult(mcpResponse);
        } else {
          // Use traditional troubleshooting mode
          const [analysis, suggestions] = await Promise.all([
            troubleshootKubernetesError({ errorMessage: values.errorMessage }),
            suggestResourceContext({ errorDescription: values.errorMessage }),
          ]);

          if (!analysis || !suggestions) {
              throw new Error("Failed to get a response from the AI assistant.");
          }
          
          setAnalysisResult(analysis);
          setSuggestedResources(suggestions);
        }

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

  const VerboseLog = ({ children }: { children: React.ReactNode }) => {
    if (!verbosity || !isPending) return null;
    return (
        <div className="font-code text-xs text-muted-foreground p-3 bg-muted/50 rounded-md my-4 animate-pulse">
            <p>&gt; {children}</p>
        </div>
    )
  }

  const renderLoadingState = () => (
    <div className="mt-8 space-y-6">
        <VerboseLog>
          {queryMode === 'natural' 
            ? 'Processing natural language query and checking if live cluster data is needed...'
            : 'Sending error description to AI investigator...'
          }
        </VerboseLog>
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Skeleton className="h-6 w-6 rounded-full" />
                    <Skeleton className="h-6 w-1/2" />
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-4/5" />
                <Skeleton className="h-4 w-full" />
            </CardContent>
        </Card>
        <Card>
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <Skeleton className="h-6 w-6 rounded-full" />
                    <Skeleton className="h-6 w-2/3" />
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-4/5" />
            </CardContent>
        </Card>
    </div>
  );

  return (
    <div className="max-w-4xl mx-auto">
      <Card className="shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Wand2 className="text-primary" />
            New Investigation
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="errorMessage"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {queryMode === 'troubleshoot' 
                        ? 'Kubernetes Error Message or Log' 
                        : 'Natural Language Query'
                      }
                    </FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={
                          queryMode === 'troubleshoot'
                            ? "e.g., 'Error: ImagePullBackOff' or paste your kubectl logs here..."
                            : "e.g., 'What is the health of my pods in kube-system namespace?' or 'Show me recent events in default namespace'"
                        }
                        className="min-h-[120px] font-code text-base"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              {/* Query Mode Selector */}
              <div className="flex items-center justify-center space-x-4 p-4 bg-muted/30 rounded-lg">
                <div className="flex items-center space-x-2">
                  <input
                    type="radio"
                    id="troubleshoot-mode"
                    name="queryMode"
                    checked={queryMode === 'troubleshoot'}
                    onChange={() => setQueryMode('troubleshoot')}
                    className="sr-only"
                  />
                  <label
                    htmlFor="troubleshoot-mode"
                    className={`flex items-center space-x-2 px-4 py-2 rounded-md cursor-pointer transition-all ${
                      queryMode === 'troubleshoot'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-background hover:bg-muted'
                    }`}
                  >
                    <Wrench className="h-4 w-4" />
                    <span>Troubleshoot Errors</span>
                  </label>
                </div>
                <div className="flex items-center space-x-2">
                  <input
                    type="radio"
                    id="natural-mode"
                    name="queryMode"
                    checked={queryMode === 'natural'}
                    onChange={() => setQueryMode('natural')}
                    className="sr-only"
                  />
                  <label
                    htmlFor="natural-mode"
                    className={`flex items-center space-x-2 px-4 py-2 rounded-md cursor-pointer transition-all ${
                      queryMode === 'natural'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-background hover:bg-muted'
                    }`}
                  >
                    <BrainCircuit className="h-4 w-4" />
                    <span>Natural Language Query</span>
                  </label>
                </div>
              </div>

              <div className="flex items-center justify-between flex-wrap gap-4">
                <div className="flex items-center space-x-2">
                  <Switch id="verbosity-switch" checked={verbosity} onCheckedChange={setVerbosity} />
                  <Label htmlFor="verbosity-switch">Verbose Mode</Label>
                </div>
                <Button type="submit" disabled={isPending}>
                  {isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Investigating...
                    </>
                  ) : (
                    <>
                      <Sparkles className="mr-2 h-4 w-4" />
                      Start Analysis
                    </>
                  )}
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
      
      {isPending && renderLoadingState()}

      {analysisResult && suggestedResources && !isPending && (
        <div className="mt-8 space-y-6 animate-in fade-in duration-500">
            <Card className="border-primary/50">
              <CardHeader>
                <CardTitle className="flex items-center gap-3 text-primary">
                    <Lightbulb />
                    Potential Causes
                </CardTitle>
              </CardHeader>
              <CardContent>
                  <ul className="space-y-3 font-code text-sm list-disc list-inside">
                      {analysisResult.potentialCauses.map((cause, i) => <li key={i}>{cause}</li>)}
                  </ul>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-3">
                    <Wrench />
                    Suggested Solutions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3 font-code text-sm list-disc list-inside">
                    {analysisResult.suggestedSolutions.map((solution, i) => <li key={i}>{solution}</li>)}
                </ul>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-3">
                    <FileText />
                    Recommended Context
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                    <div>
                        <h4 className="font-semibold mb-2 flex items-center gap-2">
                            <BrainCircuit className="w-5 h-5 text-primary/80"/>
                            Reasoning
                        </h4>
                        <p className="text-muted-foreground text-sm">{suggestedResources.reasoning}</p>
                    </div>
                    <div>
                        <h4 className="font-semibold mb-2">Suggested Resources:</h4>
                        <div className="flex flex-wrap gap-2">
                            {suggestedResources.suggestedResources.map((resource, i) => (
                                <code key={i} className="px-2 py-1 bg-accent/20 rounded-md text-sm text-accent-foreground border border-accent/30">{resource}</code>
                            ))}
                        </div>
                    </div>
                </div>
              </CardContent>
            </Card>
        </div>
      )}

      {mcpResult && !isPending && (
        <div className="mt-8 space-y-6 animate-in fade-in duration-500">
          <Card className="border-primary/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-3 text-primary">
                <BrainCircuit />
                AI Analysis
                {mcpResult.usedTool && (
                  <span className="text-xs bg-primary/10 text-primary px-2 py-1 rounded-full">
                    Used: {mcpResult.toolUsed}
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="prose prose-sm max-w-none dark:prose-invert text-sm leading-relaxed">
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  rehypePlugins={[rehypeHighlight]}
                  components={{
                    // Custom styling for better integration
                    h1: ({node, ...props}) => <h1 className="text-lg font-bold mt-4 mb-2 text-primary" {...props} />,
                    h2: ({node, ...props}) => <h2 className="text-base font-semibold mt-3 mb-2 text-primary" {...props} />,
                    h3: ({node, ...props}) => <h3 className="text-sm font-medium mt-2 mb-1" {...props} />,
                    ul: ({node, ...props}) => <ul className="space-y-1 ml-4" {...props} />,
                    ol: ({node, ...props}) => <ol className="space-y-1 ml-4" {...props} />,
                    li: ({node, ...props}) => <li className="text-sm" {...props} />,
                    p: ({node, ...props}) => <p className="mb-2 last:mb-0" {...props} />,
                    code: ({node, className, children, ...props}) => {
                      const match = /language-(\w+)/.exec(className || '');
                      return match ? (
                        <code className={`${className} text-xs`} {...props}>
                          {children}
                        </code>
                      ) : (
                        <code className="bg-muted px-1 py-0.5 rounded text-xs font-mono" {...props}>
                          {children}
                        </code>
                      );
                    },
                    pre: ({node, ...props}) => (
                      <pre className="bg-muted p-3 rounded-md overflow-x-auto text-xs" {...props} />
                    ),
                    strong: ({node, ...props}) => <strong className="font-semibold text-foreground" {...props} />,
                    blockquote: ({node, ...props}) => (
                      <blockquote className="border-l-4 border-primary/30 pl-3 italic text-muted-foreground" {...props} />
                    ),
                  }}
                >
                  {mcpResult.response}
                </ReactMarkdown>
              </div>
              
              {mcpResult.error && (
                <div className="mt-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                  <p className="text-destructive text-sm">
                    <strong>Error:</strong> {mcpResult.error}
                  </p>
                </div>
              )}
              
              {mcpResult.rawData && verbosity && (
                <details className="mt-4 p-3 bg-muted/50 rounded-md">
                  <summary className="cursor-pointer text-sm font-medium text-muted-foreground mb-2">
                    View Raw Cluster Data
                  </summary>
                  <pre className="text-xs overflow-x-auto whitespace-pre-wrap text-muted-foreground">
                    {mcpResult.rawData}
                  </pre>
                </details>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
