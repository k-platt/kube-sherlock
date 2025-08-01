package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"kube-sherlock/internal/ai"
	"kube-sherlock/internal/config"
	"kube-sherlock/internal/kubernetes"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [error-message]",
	Short: "Analyze a Kubernetes error and get troubleshooting suggestions",
	Long: `Analyze a Kubernetes error message and get AI-powered troubleshooting suggestions.
You can provide the error message as an argument or pipe it via stdin.

Examples:
  kube-sherlock analyze "ImagePullBackOff"
  kubectl logs pod/failing-pod | kube-sherlock analyze
  kube-sherlock analyze --gather-resources --namespace default "CrashLoopBackOff"`,
	Args: cobra.MaximumNArgs(1),
	Run:  runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().String("gemini-api-key", "", "Google AI (Gemini) API key")
	analyzeCmd.Flags().BoolP("gather-resources", "g", false, "Gather related Kubernetes resources for additional context")
	analyzeCmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace to gather resources from")
	analyzeCmd.Flags().StringSlice("resource-types", []string{"pods", "deployments", "services", "events"}, "Types of resources to gather")
	analyzeCmd.Flags().String("label-selector", "", "Label selector for filtering resources")
	analyzeCmd.Flags().BoolP("verbose-output", "V", false, "Show detailed analysis steps")

	viper.BindPFlag("gemini.api_key", analyzeCmd.Flags().Lookup("gemini-api-key"))
	viper.BindPFlag("gather.resources", analyzeCmd.Flags().Lookup("gather-resources"))
	viper.BindPFlag("gather.namespace", analyzeCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("gather.resource_types", analyzeCmd.Flags().Lookup("resource-types"))
	viper.BindPFlag("gather.label_selector", analyzeCmd.Flags().Lookup("label-selector"))
	viper.BindPFlag("output.verbose", analyzeCmd.Flags().Lookup("verbose-output"))
}

func runAnalyze(cmd *cobra.Command, args []string) {
	cfg := config.GetConfig()
	logger := config.GetLogger()

	// Validate required configuration
	if cfg.Gemini.APIKey == "" {
		fmt.Fprintf(os.Stderr, "Error: Gemini API key is required. Set via --gemini-api-key flag or GEMINI_API_KEY environment variable\n")
		os.Exit(1)
	}

	// Get error message from args or stdin
	var errorMessage string
	if len(args) > 0 {
		errorMessage = args[0]
	} else {
		// Read from stdin
		fmt.Fprintf(os.Stderr, "Reading error message from stdin...\n")
		// For now, require explicit error message
		fmt.Fprintf(os.Stderr, "Error: Please provide an error message as an argument\n")
		os.Exit(1)
	}

	if errorMessage == "" {
		fmt.Fprintf(os.Stderr, "Error: No error message provided\n")
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize AI service
	aiService := ai.NewService(cfg.Gemini.APIKey, cfg.Gemini.Model, logger)
	defer aiService.Close()

	verboseOutput := viper.GetBool("output.verbose")

	// Print header
	fmt.Println("🔍 Kube Sherlock Analysis")
	fmt.Println("=" + fmt.Sprintf("%*s", 24, ""))
	fmt.Printf("Error: %s\n\n", errorMessage)

	if verboseOutput {
		fmt.Println("📋 Starting AI analysis...")
	}

	// Step 1: Troubleshoot the error
	troubleshootResp, err := aiService.TroubleshootError(ctx, errorMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing error message: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Get resource suggestions
	suggestResp, err := aiService.SuggestResources(ctx, errorMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting resource suggestions: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Gather resources if requested
	var resourceContext string
	if viper.GetBool("gather.resources") {
		if verboseOutput {
			fmt.Println("📦 Gathering Kubernetes resources...")
		}

		k8sService, err := kubernetes.NewService(cfg.Kubernetes.ConfigPath, cfg.Kubernetes.Context, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to connect to Kubernetes cluster: %v\n", err)
		} else {
			namespace := viper.GetString("gather.namespace")
			resourceTypes := viper.GetStringSlice("gather.resource_types")
			labelSelector := viper.GetString("gather.label_selector")

			resources, err := k8sService.GatherResources(ctx, resourceTypes, namespace, labelSelector)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to gather resources: %v\n", err)
			} else {
				// Summarize the gathered resources
				resourceData := fmt.Sprintf("%+v", resources)
				summaryResp, err := aiService.SummarizeResourceData(ctx, resourceData)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to summarize resource data: %v\n", err)
				} else {
					resourceContext = summaryResp.Summary
				}
			}
		}
	}

	// Display results
	fmt.Println("💡 Potential Causes:")
	fmt.Println(strings.Repeat("-", 20))
	for i, cause := range troubleshootResp.PotentialCauses {
		fmt.Printf("%d. %s\n", i+1, cause)
	}

	fmt.Println("\n🔧 Suggested Solutions:")
	fmt.Println(strings.Repeat("-", 23))
	for i, solution := range troubleshootResp.SuggestedSolutions {
		fmt.Printf("%d. %s\n", i+1, solution)
	}

	fmt.Println("\n📋 Recommended Resources to Check:")
	fmt.Println(strings.Repeat("-", 37))
	fmt.Printf("Reasoning: %s\n\n", suggestResp.Reasoning)
	for i, resource := range suggestResp.SuggestedResources {
		fmt.Printf("%d. %s\n", i+1, resource)
	}

	if resourceContext != "" {
		fmt.Println("\n📊 Current Cluster Context:")
		fmt.Println(strings.Repeat("-", 28))
		fmt.Println(resourceContext)
	}

	if verboseOutput {
		fmt.Println("\n✅ Analysis complete!")
	}
}
