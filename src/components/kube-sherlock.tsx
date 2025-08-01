"use client";

import { useState, useTransition } from "react";
import { troubleshootKubernetesError, type TroubleshootKubernetesErrorOutput } from "@/ai/flows/troubleshoot-kubernetes-error";
import { suggestResourceContext, type SuggestResourceContextOutput } from "@/ai/flows/suggest-resource-context";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

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
    message: "Please enter a more detailed error message (at least 10 characters).",
  }),
});

export function KubeSherlock() {
  const [isPending, startTransition] = useTransition();
  const [verbosity, setVerbosity] = useState(false);
  const [analysisResult, setAnalysisResult] = useState<TroubleshootKubernetesErrorOutput | null>(null);
  const [suggestedResources, setSuggestedResources] = useState<SuggestResourceContextOutput | null>(null);
  const { toast } = useToast();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      errorMessage: "",
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    setAnalysisResult(null);
    setSuggestedResources(null);

    startTransition(async () => {
      try {
        if (verbosity) {
          console.log("Starting investigation with error: ", values.errorMessage);
        }
        const [analysis, suggestions] = await Promise.all([
          troubleshootKubernetesError({ errorMessage: values.errorMessage }),
          suggestResourceContext({ errorDescription: values.errorMessage }),
        ]);

        if (!analysis || !suggestions) {
            throw new Error("Failed to get a response from the AI assistant.");
        }
        
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
        <VerboseLog>Sending error description to AI investigator...</VerboseLog>
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
                    <FormLabel>Kubernetes Error Message or Log</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="e.g., 'Error: ImagePullBackOff' or paste your kubectl logs here..."
                        className="min-h-[120px] font-code text-base"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
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
    </div>
  );
}
