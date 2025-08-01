package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Gemini     GeminiConfig     `mapstructure:"gemini"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type GeminiConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	Context    string `mapstructure:"context"`
}

var (
	globalConfig *Config
	globalLogger *zap.Logger
)

// GetConfig returns the global configuration
func GetConfig() *Config {
	if globalConfig == nil {
		globalConfig = &Config{
			Server: ServerConfig{
				Host: viper.GetString("server.host"),
				Port: viper.GetString("server.port"),
			},
			Gemini: GeminiConfig{
				APIKey: viper.GetString("gemini.api_key"),
				Model:  viper.GetString("gemini.model"),
			},
			Kubernetes: KubernetesConfig{
				ConfigPath: viper.GetString("kubernetes.config_path"),
				Context:    viper.GetString("kubernetes.context"),
			},
		}

		// Set defaults
		if globalConfig.Server.Host == "" {
			globalConfig.Server.Host = "localhost"
		}
		if globalConfig.Server.Port == "" {
			globalConfig.Server.Port = "8080"
		}
		if globalConfig.Gemini.Model == "" {
			globalConfig.Gemini.Model = "gemini-2.0-flash"
		}
	}
	return globalConfig
}

// SetLogger sets the global logger
func SetLogger(logger *zap.Logger) {
	globalLogger = logger
}

// GetLogger returns the global logger
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// Fallback logger
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger
}
