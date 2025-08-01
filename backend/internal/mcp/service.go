package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"kube-sherlock/internal/kubernetes"

	"go.uber.org/zap"
)

// Tool represents an MCP tool that can be executed
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

// ToolSchema defines the input parameters for a tool
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

// ToolResult represents the result of tool execution
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content returned by a tool
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolRequest represents a request to execute a tool
type ToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPService handles Model Context Protocol operations
type MCPService struct {
	k8sService *kubernetes.Service
	logger     *zap.Logger
	tools      map[string]Tool
}

// NewMCPService creates a new MCP service
func NewMCPService(k8sService *kubernetes.Service, logger *zap.Logger) *MCPService {
	mcp := &MCPService{
		k8sService: k8sService,
		logger:     logger,
		tools:      make(map[string]Tool),
	}

	// Register built-in tools
	mcp.registerTools()
	return mcp
}

// registerTools registers all available MCP tools
func (m *MCPService) registerTools() {
	// Get pod health tool
	m.tools["get_pod_health"] = Tool{
		Name:        "get_pod_health",
		Description: "Get the health status of pods in a namespace",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Kubernetes namespace to check (default: default)",
				},
				"labelSelector": map[string]interface{}{
					"type":        "string",
					"description": "Label selector to filter pods (optional)",
				},
			},
			Required: []string{},
		},
	}

	// Get deployment status tool
	m.tools["get_deployment_status"] = Tool{
		Name:        "get_deployment_status",
		Description: "Get the status of deployments in a namespace",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Kubernetes namespace to check (default: default)",
				},
				"deploymentName": map[string]interface{}{
					"type":        "string",
					"description": "Specific deployment name (optional)",
				},
			},
			Required: []string{},
		},
	}

	// Get service endpoints tool
	m.tools["get_service_endpoints"] = Tool{
		Name:        "get_service_endpoints",
		Description: "Get the endpoints and status of services in a namespace",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Kubernetes namespace to check (default: default)",
				},
				"serviceName": map[string]interface{}{
					"type":        "string",
					"description": "Specific service name (optional)",
				},
			},
			Required: []string{},
		},
	}

	// Get recent events tool
	m.tools["get_recent_events"] = Tool{
		Name:        "get_recent_events",
		Description: "Get recent Kubernetes events in a namespace",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Kubernetes namespace to check (default: default)",
				},
				"resourceName": map[string]interface{}{
					"type":        "string",
					"description": "Filter events for specific resource (optional)",
				},
			},
			Required: []string{},
		},
	}

	// Get pod logs tool
	m.tools["get_pod_logs"] = Tool{
		Name:        "get_pod_logs",
		Description: "Get logs from a specific pod",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":        "string",
					"description": "Kubernetes namespace (default: default)",
				},
				"podName": map[string]interface{}{
					"type":        "string",
					"description": "Name of the pod to get logs from",
				},
				"containerName": map[string]interface{}{
					"type":        "string",
					"description": "Container name (optional)",
				},
				"lines": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to retrieve (default: 100)",
				},
			},
			Required: []string{"podName"},
		},
	}
}

// ListTools returns all available tools
func (m *MCPService) ListTools() []Tool {
	tools := make([]Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool executes a specific tool with given arguments
func (m *MCPService) ExecuteTool(ctx context.Context, request ToolRequest) (*ToolResult, error) {
	_, exists := m.tools[request.Name]
	if !exists {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Unknown tool: %s", request.Name),
			}},
			IsError: true,
		}, fmt.Errorf("unknown tool: %s", request.Name)
	}

	m.logger.Info("Executing MCP tool",
		zap.String("tool", request.Name),
		zap.Any("arguments", request.Arguments))

	switch request.Name {
	case "get_pod_health":
		return m.getPodHealth(ctx, request.Arguments)
	case "get_deployment_status":
		return m.getDeploymentStatus(ctx, request.Arguments)
	case "get_service_endpoints":
		return m.getServiceEndpoints(ctx, request.Arguments)
	case "get_recent_events":
		return m.getRecentEvents(ctx, request.Arguments)
	case "get_pod_logs":
		return m.getPodLogs(ctx, request.Arguments)
	default:
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Tool %s is registered but not implemented", request.Name),
			}},
			IsError: true,
		}, fmt.Errorf("tool not implemented: %s", request.Name)
	}
}

// Helper function to get string parameter with default
func getStringParam(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// Helper function to get int parameter with default
func getIntParam(args map[string]interface{}, key string, defaultValue int64) int64 {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int:
			return int64(v)
		case int64:
			return v
		}
	}
	return defaultValue
}

// getPodHealth retrieves pod health information
func (m *MCPService) getPodHealth(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := getStringParam(args, "namespace", "default")
	labelSelector := getStringParam(args, "labelSelector", "")

	if m.k8sService == nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Kubernetes service not available. Please ensure cluster connectivity.",
			}},
			IsError: true,
		}, fmt.Errorf("kubernetes service not available")
	}

	// Gather pod information
	resources, err := m.k8sService.GatherResources(ctx, []string{"pods"}, namespace, labelSelector)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error gathering pod information: %v", err),
			}},
			IsError: true,
		}, err
	}

	// Format the response
	podsData, _ := json.MarshalIndent(resources.Resources["pods"], "", "  ")

	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Pod health information for namespace '%s':\n\n%s", namespace, string(podsData)),
		}},
	}, nil
}

// getDeploymentStatus retrieves deployment status information
func (m *MCPService) getDeploymentStatus(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := getStringParam(args, "namespace", "default")
	deploymentName := getStringParam(args, "deploymentName", "")

	if m.k8sService == nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Kubernetes service not available. Please ensure cluster connectivity.",
			}},
			IsError: true,
		}, fmt.Errorf("kubernetes service not available")
	}

	labelSelector := ""
	if deploymentName != "" {
		labelSelector = fmt.Sprintf("app=%s", deploymentName)
	}

	resources, err := m.k8sService.GatherResources(ctx, []string{"deployments"}, namespace, labelSelector)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error gathering deployment information: %v", err),
			}},
			IsError: true,
		}, err
	}

	deploymentsData, _ := json.MarshalIndent(resources.Resources["deployments"], "", "  ")

	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Deployment status for namespace '%s':\n\n%s", namespace, string(deploymentsData)),
		}},
	}, nil
}

// getServiceEndpoints retrieves service endpoint information
func (m *MCPService) getServiceEndpoints(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := getStringParam(args, "namespace", "default")
	serviceName := getStringParam(args, "serviceName", "")

	if m.k8sService == nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Kubernetes service not available. Please ensure cluster connectivity.",
			}},
			IsError: true,
		}, fmt.Errorf("kubernetes service not available")
	}

	labelSelector := ""
	if serviceName != "" {
		labelSelector = fmt.Sprintf("app=%s", serviceName)
	}

	resources, err := m.k8sService.GatherResources(ctx, []string{"services"}, namespace, labelSelector)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error gathering service information: %v", err),
			}},
			IsError: true,
		}, err
	}

	servicesData, _ := json.MarshalIndent(resources.Resources["services"], "", "  ")

	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Service endpoints for namespace '%s':\n\n%s", namespace, string(servicesData)),
		}},
	}, nil
}

// getRecentEvents retrieves recent Kubernetes events
func (m *MCPService) getRecentEvents(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := getStringParam(args, "namespace", "default")
	resourceName := getStringParam(args, "resourceName", "")

	if m.k8sService == nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Kubernetes service not available. Please ensure cluster connectivity.",
			}},
			IsError: true,
		}, fmt.Errorf("kubernetes service not available")
	}

	labelSelector := ""
	if resourceName != "" {
		labelSelector = fmt.Sprintf("involvedObject.name=%s", resourceName)
	}

	resources, err := m.k8sService.GatherResources(ctx, []string{"events"}, namespace, labelSelector)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error gathering events: %v", err),
			}},
			IsError: true,
		}, err
	}

	eventsData, _ := json.MarshalIndent(resources.Resources["events"], "", "  ")

	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Recent events for namespace '%s':\n\n%s", namespace, string(eventsData)),
		}},
	}, nil
}

// getPodLogs retrieves logs from a specific pod
func (m *MCPService) getPodLogs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := getStringParam(args, "namespace", "default")
	podName := getStringParam(args, "podName", "")
	containerName := getStringParam(args, "containerName", "")
	lines := getIntParam(args, "lines", 100)

	if podName == "" {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Pod name is required for getting logs",
			}},
			IsError: true,
		}, fmt.Errorf("pod name is required")
	}

	if m.k8sService == nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Kubernetes service not available. Please ensure cluster connectivity.",
			}},
			IsError: true,
		}, fmt.Errorf("kubernetes service not available")
	}

	logs, err := m.k8sService.GetPodLogs(ctx, namespace, podName, containerName, lines)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting pod logs: %v", err),
			}},
			IsError: true,
		}, err
	}

	return &ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Logs for pod '%s' in namespace '%s' (last %d lines):\n\n%s",
				podName, namespace, lines, logs),
		}},
	}, nil
}
