package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-sherlock/internal/ai"
	"kube-sherlock/internal/kubernetes"
)

// Handler contains the API handlers and dependencies
type Handler struct {
	aiService  *ai.Service
	k8sService *kubernetes.Service
	logger     *zap.Logger
}

// TroubleshootRequest represents the request to troubleshoot a Kubernetes error
type TroubleshootRequest struct {
	ErrorMessage string `json:"errorMessage" binding:"required"`
}

// TroubleshootResponse represents the response from troubleshooting
type TroubleshootResponse struct {
	PotentialCauses    []string `json:"potentialCauses"`
	SuggestedSolutions []string `json:"suggestedSolutions"`
}

// SuggestResourcesRequest represents the request to suggest Kubernetes resources
type SuggestResourcesRequest struct {
	ErrorDescription string `json:"errorDescription" binding:"required"`
}

// SuggestResourcesResponse represents the response with suggested resources
type SuggestResourcesResponse struct {
	SuggestedResources []string `json:"suggestedResources"`
	Reasoning          string   `json:"reasoning"`
}

// SummarizeRequest represents the request to summarize resource data
type SummarizeRequest struct {
	ResourceData string `json:"resourceData" binding:"required"`
}

// SummarizeResponse represents the response with summarized data
type SummarizeResponse struct {
	Summary string `json:"summary"`
}

// GatherResourcesRequest represents the request to gather Kubernetes resources
type GatherResourcesRequest struct {
	ResourceTypes []string `json:"resourceTypes" binding:"required"`
	Namespace     string   `json:"namespace"`
	LabelSelector string   `json:"labelSelector"`
}

// GatherResourcesResponse represents the response with gathered resource data
type GatherResourcesResponse struct {
	Resources map[string]interface{} `json:"resources"`
	Metadata  GatherMetadata         `json:"metadata"`
}

// GatherMetadata contains metadata about the gathering operation
type GatherMetadata struct {
	Timestamp      string `json:"timestamp"`
	ClusterContext string `json:"clusterContext"`
	Namespace      string `json:"namespace"`
}

// MCPQueryRequest represents a natural language query request
type MCPQueryRequest struct {
	Query string `json:"query" binding:"required"`
}

// MCPQueryResponse represents the response from an MCP query
type MCPQueryResponse struct {
	Response string `json:"response"`
	UsedTool bool   `json:"usedTool"`
	ToolUsed string `json:"toolUsed,omitempty"`
	RawData  string `json:"rawData,omitempty"`
	Error    string `json:"error,omitempty"`
}

// health is a simple health check endpoint
func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "kube-sherlock",
	})
}

// troubleshoot handles Kubernetes error troubleshooting requests
func (h *Handler) troubleshoot(c *gin.Context) {
	var req TroubleshootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid troubleshoot request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Processing troubleshoot request", zap.String("error", req.ErrorMessage))

	response, err := h.aiService.TroubleshootError(c.Request.Context(), req.ErrorMessage)
	if err != nil {
		h.logger.Error("Failed to troubleshoot error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// suggestResources handles resource suggestion requests
func (h *Handler) suggestResources(c *gin.Context) {
	var req SuggestResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid suggest resources request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Processing suggest resources request", zap.String("description", req.ErrorDescription))

	response, err := h.aiService.SuggestResources(c.Request.Context(), req.ErrorDescription)
	if err != nil {
		h.logger.Error("Failed to suggest resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suggest resources"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// summarize handles resource data summarization requests
func (h *Handler) summarize(c *gin.Context) {
	var req SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid summarize request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Processing summarize request")

	response, err := h.aiService.SummarizeResourceData(c.Request.Context(), req.ResourceData)
	if err != nil {
		h.logger.Error("Failed to summarize resource data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to summarize data"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// gatherResources handles Kubernetes resource gathering requests
func (h *Handler) gatherResources(c *gin.Context) {
	var req GatherResourcesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid gather resources request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.k8sService == nil {
		h.logger.Error("Kubernetes service not available")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Kubernetes service not configured"})
		return
	}

	h.logger.Info("Processing gather resources request",
		zap.Strings("types", req.ResourceTypes),
		zap.String("namespace", req.Namespace))

	response, err := h.k8sService.GatherResources(c.Request.Context(), req.ResourceTypes, req.Namespace, req.LabelSelector)
	if err != nil {
		h.logger.Error("Failed to gather resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to gather resources"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// mcpQuery handles natural language queries with MCP tool support
func (h *Handler) mcpQuery(c *gin.Context) {
	var req MCPQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid MCP query request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Processing MCP query", zap.String("query", req.Query))

	response, err := h.aiService.QueryWithMCP(c.Request.Context(), req.Query)
	if err != nil {
		h.logger.Error("Failed to process MCP query", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query"})
		return
	}

	c.JSON(http.StatusOK, response)
}
