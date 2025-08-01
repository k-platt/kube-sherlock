package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Service handles Kubernetes cluster interactions
type Service struct {
	clientset   *kubernetes.Clientset
	config      *rest.Config
	contextName string
	logger      *zap.Logger
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

// NewService creates a new Kubernetes service
func NewService(configPath, contextName string, logger *zap.Logger) (*Service, error) {
	var config *rest.Config
	var err error

	if configPath == "" {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to kubeconfig
			if home := homedir.HomeDir(); home != "" {
				configPath = filepath.Join(home, ".kube", "config")
			}
		}
	}

	if config == nil {
		// Load from kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			logger.Error("Failed to load kubeconfig", zap.Error(err), zap.String("path", configPath))
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("Failed to create Kubernetes clientset", zap.Error(err))
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Test connection
	testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = clientset.CoreV1().Namespaces().List(testCtx, metav1.ListOptions{Limit: 1})
	if err != nil {
		logger.Error("Failed to connect to Kubernetes cluster", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to cluster: %w", err)
	}

	logger.Info("Successfully connected to Kubernetes cluster")

	return &Service{
		clientset:   clientset,
		config:      config,
		contextName: contextName,
		logger:      logger,
	}, nil
}

// GatherResources gathers specified Kubernetes resources
func (s *Service) GatherResources(ctx context.Context, resourceTypes []string, namespace, labelSelector string) (*GatherResourcesResponse, error) {
	resources := make(map[string]interface{})

	// If no namespace specified, use "default"
	if namespace == "" {
		namespace = "default"
	}

	s.logger.Info("Gathering resources",
		zap.Strings("types", resourceTypes),
		zap.String("namespace", namespace),
		zap.String("labelSelector", labelSelector))

	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	for _, resourceType := range resourceTypes {
		switch resourceType {
		case "pods":
			pods, err := s.clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list pods", zap.Error(err))
				resources["pods_error"] = err.Error()
			} else {
				resources["pods"] = pods
			}

		case "deployments":
			deployments, err := s.clientset.AppsV1().Deployments(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list deployments", zap.Error(err))
				resources["deployments_error"] = err.Error()
			} else {
				resources["deployments"] = deployments
			}

		case "services":
			services, err := s.clientset.CoreV1().Services(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list services", zap.Error(err))
				resources["services_error"] = err.Error()
			} else {
				resources["services"] = services
			}

		case "configmaps":
			configMaps, err := s.clientset.CoreV1().ConfigMaps(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list configmaps", zap.Error(err))
				resources["configmaps_error"] = err.Error()
			} else {
				resources["configmaps"] = configMaps
			}

		case "secrets":
			secrets, err := s.clientset.CoreV1().Secrets(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list secrets", zap.Error(err))
				resources["secrets_error"] = err.Error()
			} else {
				// Redact secret data for security
				for i := range secrets.Items {
					secrets.Items[i].Data = map[string][]byte{}
					secrets.Items[i].StringData = map[string]string{}
				}
				resources["secrets"] = secrets
			}

		case "events":
			events, err := s.clientset.CoreV1().Events(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list events", zap.Error(err))
				resources["events_error"] = err.Error()
			} else {
				resources["events"] = events
			}

		case "replicasets":
			replicaSets, err := s.clientset.AppsV1().ReplicaSets(namespace).List(ctx, listOptions)
			if err != nil {
				s.logger.Error("Failed to list replicasets", zap.Error(err))
				resources["replicasets_error"] = err.Error()
			} else {
				resources["replicasets"] = replicaSets
			}

		default:
			s.logger.Warn("Unsupported resource type", zap.String("type", resourceType))
			resources[resourceType+"_error"] = fmt.Sprintf("unsupported resource type: %s", resourceType)
		}
	}

	response := &GatherResourcesResponse{
		Resources: resources,
		Metadata: GatherMetadata{
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			ClusterContext: s.contextName,
			Namespace:      namespace,
		},
	}

	return response, nil
}

// GetPodLogs retrieves logs from a specific pod
func (s *Service) GetPodLogs(ctx context.Context, namespace, podName, containerName string, lines int64) (string, error) {
	options := &v1.PodLogOptions{
		Container: containerName,
	}

	if lines > 0 {
		options.TailLines = &lines
	}

	request := s.clientset.CoreV1().Pods(namespace).GetLogs(podName, options)
	logs, err := request.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 2048)
	var result []byte
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return string(result), nil
}
