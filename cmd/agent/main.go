package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"

	"github.com/rizesky/mckmt/internal/agent"
	"github.com/rizesky/mckmt/internal/config"
	"github.com/rizesky/mckmt/internal/kube"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting MCKMA Agent", zap.String("version", "1.0.0"))

	// Initialize Kubernetes client
	kubeClient, err := initKubeClient(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kubernetes client", zap.Error(err))
	}

	// Initialize agent
	agent := agent.NewAgent(cfg, kubeClient, logger)

	// Note: Cluster ID will be assigned by the hub during registration

	// Start agent
	if err := agent.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Agent shutting down...")
	agent.Stop()
}

// initKubeClient initializes the Kubernetes client
func initKubeClient(cfg *config.Config, logger *zap.Logger) (*kube.Client, error) {
	var kubeconfig []byte
	var err error

	// Check if we're running in-cluster (inside a Kubernetes pod)
	if _, err := rest.InClusterConfig(); err == nil {
		logger.Info("Using in-cluster Kubernetes configuration")
		// Pass empty kubeconfig to use in-cluster config
		kubeconfig = []byte{}
	} else {
		// Not running in-cluster, try to load kubeconfig from various sources
		logger.Info("Loading kubeconfig from file sources")
		kubeconfig, err = loadKubeconfigFromSources()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	client, err := kube.NewClient(kubeconfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client, nil
}

// loadKubeconfigFromSources tries to load kubeconfig from various sources
func loadKubeconfigFromSources() ([]byte, error) {
	// 1. Check for kubeconfig content in environment variable
	if kubeconfigContent := os.Getenv("KUBECONFIG_CONTENT"); kubeconfigContent != "" {
		return []byte(kubeconfigContent), nil
	}

	// 2. Check for kubeconfig file path in environment variable
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		if content, err := ioutil.ReadFile(kubeconfigPath); err == nil {
			return content, nil
		}
	}

	// 3. Check for kubeconfig file path in MCKMA_KUBECONFIG_PATH
	if kubeconfigPath := os.Getenv("MCKMA_KUBECONFIG_PATH"); kubeconfigPath != "" {
		if content, err := ioutil.ReadFile(kubeconfigPath); err == nil {
			return content, nil
		}
	}

	// 4. Try default kubeconfig locations
	defaultPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"/etc/kubernetes/admin.conf",
		"/etc/kubernetes/kubelet.conf",
		"./kubeconfig",
	}

	for _, path := range defaultPaths {
		if content, err := ioutil.ReadFile(path); err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("no kubeconfig found in any of the expected locations")
}

// initLogger initializes the logger based on configuration
func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(getLogLevel(cfg.Level))
	config.DisableCaller = !cfg.Caller
	config.DisableStacktrace = !cfg.Stacktrace

	return config.Build()
}

// getLogLevel converts string level to zap level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
