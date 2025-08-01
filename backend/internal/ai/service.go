package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/option"

	"kube-sherlock/internal/mcp"
)

// Service handles AI-powered analysis using Google Gemini
type Service struct {
	client     *genai.Client
	model      string
	logger     *zap.Logger
	mcpService *mcp.MCPService
}

// TroubleshootResponse represents the response from troubleshooting
type TroubleshootResponse struct {
	PotentialCauses    []string `json:"potentialCauses"`
	SuggestedSolutions []string `json:"suggestedSolutions"`
}

// SuggestResourcesResponse represents the response with suggested resources
type SuggestResourcesResponse struct {
	SuggestedResources []string `json:"suggestedResources"`
	Reasoning          string   `json:"reasoning"`
}

// SummarizeResponse represents the response with summarized data
type SummarizeResponse struct {
	Summary string `json:"summary"`
}

// NewService creates a new AI service
func NewService(apiKey, model string, logger *zap.Logger) *Service {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.Fatal("Failed to create Gemini client", zap.Error(err))
	}

	return &Service{
		client:     client,
		model:      model,
		logger:     logger,
		mcpService: nil, // Will be set later when needed
	}
}

// SetMCPService sets the MCP service for tool execution
func (s *Service) SetMCPService(mcpService *mcp.MCPService) {
	s.mcpService = mcpService
}

// Close closes the AI service client
func (s *Service) Close() error {
	return s.client.Close()
}

// TroubleshootError analyzes a Kubernetes error and provides troubleshooting guidance
func (s *Service) TroubleshootError(ctx context.Context, errorMessage string) (*TroubleshootResponse, error) {
	prompt := fmt.Sprintf(`You are a Kubernetes expert specializing in troubleshooting errors. Analyze the provided error message or event description to determine potential causes and suggest solutions.

Error Message/Event Description: %s

Provide your output in the following JSON format:
{
  "potentialCauses": ["cause1", "cause2", "cause3"],
  "suggestedSolutions": ["solution1", "solution2", "solution3"]
}

Focus on practical, actionable solutions. Be specific about kubectl commands, configuration changes, or diagnostic steps.`, errorMessage)

	model := s.client.GenerativeModel(s.model)
	model.SetTemperature(0.1) // Lower temperature for more consistent technical responses

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		s.logger.Error("Failed to generate content for troubleshooting", zap.Error(err))
		return nil, fmt.Errorf("failed to analyze error: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Extract text from response
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// Parse JSON response
	var result TroubleshootResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.logger.Error("Failed to parse AI response", zap.Error(err), zap.String("response", responseText))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// SuggestResources suggests Kubernetes resources for troubleshooting context
func (s *Service) SuggestResources(ctx context.Context, errorDescription string) (*SuggestResourcesResponse, error) {
	prompt := fmt.Sprintf(`You are a Kubernetes troubleshooting expert. Given the following error description, suggest which Kubernetes resources (e.g., logs from specific pods, deployment configurations) would provide helpful context for troubleshooting.

Error Description: %s

Respond with a list of suggested resources and a brief explanation of why each resource is relevant. Focus on resources that would help diagnose the root cause of the error.

Provide your output in the following JSON format:
{
  "suggestedResources": ["pod/example-pod logs", "deployment/example-deployment configuration", "service/example-service description"],
  "reasoning": "Pod logs may contain error messages. Deployment configuration can show misconfigurations. Service description can show service unavailable issues."
}`, errorDescription)

	model := s.client.GenerativeModel(s.model)
	model.SetTemperature(0.1)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		s.logger.Error("Failed to generate content for resource suggestions", zap.Error(err))
		return nil, fmt.Errorf("failed to suggest resources: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Extract text from response
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// Parse JSON response
	var result SuggestResourcesResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.logger.Error("Failed to parse AI response", zap.Error(err), zap.String("response", responseText))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// SummarizeResourceData summarizes Kubernetes resource data for diagnosis
func (s *Service) SummarizeResourceData(ctx context.Context, resourceData string) (*SummarizeResponse, error) {
	prompt := fmt.Sprintf(`You are an expert Kubernetes troubleshooter. Your task is to summarize the provided data from Kubernetes resources, highlighting only the relevant information for diagnosing issues. Ignore any irrelevant details.

Resource Data:
%s

Provide your output in the following JSON format:
{
  "summary": "A summarized version of the input resource data, highlighting the relevant information for diagnosing issues."
}`, resourceData)

	model := s.client.GenerativeModel(s.model)
	model.SetTemperature(0.1)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		s.logger.Error("Failed to generate content for summarization", zap.Error(err))
		return nil, fmt.Errorf("failed to summarize data: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Extract text from response
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// Parse JSON response
	var result SummarizeResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.logger.Error("Failed to parse AI response", zap.Error(err), zap.String("response", responseText))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// QueryWithMCP handles natural language queries with MCP tool support
func (s *Service) QueryWithMCP(ctx context.Context, query string) (*QueryResponse, error) {
	if s.mcpService == nil {
		return nil, fmt.Errorf("MCP service not available")
	}

	// Create a prompt that includes available tools
	tools := s.mcpService.ListTools()
	toolsJSON, _ := json.MarshalIndent(tools, "", "  ")

	prompt := fmt.Sprintf(`You are a Kubernetes expert assistant. Answer the user's query using available tools when needed.

Query: %s

Available Tools:
%s

CRITICAL: Respond ONLY with valid JSON. NO explanations, NO markdown, NO code blocks.

If you need cluster data:
{"action": "use_tool", "tool": "tool_name", "arguments": {"param": "value"}}

If you can answer directly:
{"action": "answer", "response": "## Your markdown-formatted answer here\n\nUse proper markdown formatting with headers, bullet points, and **bold** text for better readability."}

Choose the most appropriate tool for the query and respond immediately.`, query, string(toolsJSON))

	model := s.client.GenerativeModel(s.model)
	model.SetTemperature(0.1)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		s.logger.Error("Failed to generate MCP response", zap.Error(err))
		return nil, fmt.Errorf("failed to process query: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Extract text from response
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// Parse the AI response to see if it wants to use a tool
	var aiAction struct {
		Action    string                 `json:"action"`
		Tool      string                 `json:"tool"`
		Arguments map[string]interface{} `json:"arguments"`
		Response  string                 `json:"response"`
	}

	// Extract JSON from markdown code blocks if present
	jsonText := responseText
	if strings.Contains(responseText, "```json") {
		// Find the JSON block
		start := strings.Index(responseText, "```json") + 7
		end := strings.Index(responseText[start:], "```")
		if end != -1 {
			jsonText = strings.TrimSpace(responseText[start : start+end])
		}
	}

	// First try to parse as direct action
	if err := json.Unmarshal([]byte(jsonText), &aiAction); err != nil {
		// If that fails, check if there's a JSON block within the response text
		if strings.Contains(responseText, "{") && strings.Contains(responseText, "}") {
			// Try to extract any JSON from the response
			start := strings.Index(responseText, "{")
			end := strings.LastIndex(responseText, "}")
			if start != -1 && end != -1 && end > start {
				possibleJSON := responseText[start : end+1]
				if err := json.Unmarshal([]byte(possibleJSON), &aiAction); err == nil {
					// Successfully parsed JSON from within the response
					jsonText = possibleJSON
				} else {
					// If all parsing fails, treat it as a direct response
					return &QueryResponse{
						Response: responseText,
						UsedTool: false,
					}, nil
				}
			} else {
				// If parsing fails, treat it as a direct response
				return &QueryResponse{
					Response: responseText,
					UsedTool: false,
				}, nil
			}
		} else {
			// If parsing fails, treat it as a direct response
			return &QueryResponse{
				Response: responseText,
				UsedTool: false,
			}, nil
		}
	}

	if aiAction.Action == "use_tool" {
		// Execute the requested tool
		toolRequest := mcp.ToolRequest{
			Name:      aiAction.Tool,
			Arguments: aiAction.Arguments,
		}

		toolResult, err := s.mcpService.ExecuteTool(ctx, toolRequest)
		if err != nil {
			return &QueryResponse{
				Response: fmt.Sprintf("Error executing tool %s: %v", aiAction.Tool, err),
				UsedTool: true,
				ToolUsed: aiAction.Tool,
				Error:    err.Error(),
			}, nil
		}

		// Now ask AI to analyze the tool results
		var toolOutput string
		for _, content := range toolResult.Content {
			toolOutput += content.Text + "\n"
		}

		analysisPrompt := fmt.Sprintf(`Based on the following Kubernetes cluster data, provide a comprehensive answer to the user's query. 

Format your response using markdown for better readability:
- Use headers (## ) for main sections
- Use bullet points for lists
- Use **bold** for important information
- Use code blocks for kubectl commands or resource names
- Include specific recommendations and next steps

Original Query: %s

Cluster Data:
%s

Provide a well-structured markdown response analyzing this data with clear sections for current state, findings, and recommendations.`, query, toolOutput)

		analysisResp, err := model.GenerateContent(ctx, genai.Text(analysisPrompt))
		if err != nil {
			return &QueryResponse{
				Response: fmt.Sprintf("Gathered data but failed to analyze: %s", toolOutput),
				UsedTool: true,
				ToolUsed: aiAction.Tool,
			}, nil
		}

		// Extract analysis text
		analysisText := ""
		if len(analysisResp.Candidates) > 0 && len(analysisResp.Candidates[0].Content.Parts) > 0 {
			for _, part := range analysisResp.Candidates[0].Content.Parts {
				if text, ok := part.(genai.Text); ok {
					analysisText += string(text)
				}
			}
		}

		return &QueryResponse{
			Response: analysisText,
			UsedTool: true,
			ToolUsed: aiAction.Tool,
			RawData:  toolOutput,
		}, nil
	}

	// Direct answer without tools
	return &QueryResponse{
		Response: aiAction.Response,
		UsedTool: false,
	}, nil
}

// QueryResponse represents the response from an MCP-enabled query
type QueryResponse struct {
	Response string `json:"response"`
	UsedTool bool   `json:"usedTool"`
	ToolUsed string `json:"toolUsed,omitempty"`
	RawData  string `json:"rawData,omitempty"`
	Error    string `json:"error,omitempty"`
}
