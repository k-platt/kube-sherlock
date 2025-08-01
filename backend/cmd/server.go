package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"kube-sherlock/internal/api"
	"kube-sherlock/internal/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Kube Sherlock API server",
	Long:  `Start the HTTP API server to handle troubleshooting requests from the frontend.`,
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("port", "p", "8080", "Port to run the server on")
	serverCmd.Flags().String("host", "localhost", "Host to bind the server to")
	serverCmd.Flags().String("gemini-api-key", "", "Google AI (Gemini) API key")

	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.host", serverCmd.Flags().Lookup("host"))
	viper.BindPFlag("gemini.api_key", serverCmd.Flags().Lookup("gemini-api-key"))
}

func runServer(cmd *cobra.Command, args []string) {
	cfg := config.GetConfig()
	logger := config.GetLogger()

	// Validate required configuration
	if cfg.Gemini.APIKey == "" {
		logger.Fatal("Gemini API key is required. Set via --gemini-api-key flag or GEMINI_API_KEY environment variable")
	}

	// Set gin mode based on verbosity
	if viper.GetBool("verbose") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := api.NewRouter(cfg, logger)

	// Setup server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server startup failed", zap.Error(err))
		}
	}() // Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
