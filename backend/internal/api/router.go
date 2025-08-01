package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-sherlock/internal/ai"
	"kube-sherlock/internal/config"
	"kube-sherlock/internal/kubernetes"
	"kube-sherlock/internal/mcp"
)

// NewRouter creates and configures the API router
func NewRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Initialize services
	aiService := ai.NewService(cfg.Gemini.APIKey, cfg.Gemini.Model, logger)
	k8sService, err := kubernetes.NewService(cfg.Kubernetes.ConfigPath, cfg.Kubernetes.Context, logger)
	if err != nil {
		logger.Warn("Failed to initialize Kubernetes service", zap.Error(err))
		k8sService = nil // Service will handle nil gracefully
	}

	// Initialize MCP service if Kubernetes is available
	var mcpService *mcp.MCPService
	if k8sService != nil {
		mcpService = mcp.NewMCPService(k8sService, logger)
		aiService.SetMCPService(mcpService)
	}

	// API handlers
	handler := &Handler{
		aiService:  aiService,
		k8sService: k8sService,
		logger:     logger,
	}

	// Health check
	router.GET("/health", handler.health)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/troubleshoot", handler.troubleshoot)
		api.POST("/suggest-resources", handler.suggestResources)
		api.POST("/summarize", handler.summarize)
		api.POST("/gather-resources", handler.gatherResources)
		api.POST("/query", handler.mcpQuery) // New MCP endpoint
	}

	return router
}

// corsMiddleware adds CORS headers for frontend communication
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
