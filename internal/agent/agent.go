package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/rizesky/mckmt/api/proto/agent/v1"
	"github.com/rizesky/mckmt/internal/config"
	"github.com/rizesky/mckmt/internal/kube"
)

// Agent represents a cluster agent
type Agent struct {
	config       *config.AgentConfig
	kubeClient   *kube.Client
	logger       *zap.Logger
	conn         *grpc.ClientConn
	client       agentv1.AgentServiceClient
	clusterID    string
	sessionToken string
	stopCh       chan struct{}
	cancelOps    map[string]context.CancelFunc // operation_id -> cancel function
}

// NewAgent creates a new cluster agent
func NewAgent(cfg *config.AgentConfig, kubeClient *kube.Client, logger *zap.Logger) *Agent {
	return &Agent{
		config:     cfg,
		kubeClient: kubeClient,
		logger:     logger,
		stopCh:     make(chan struct{}),
		cancelOps:  make(map[string]context.CancelFunc),
	}
}

// SetClusterID sets the cluster ID for the agent
func (a *Agent) SetClusterID(clusterID string) {
	a.clusterID = clusterID
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	a.logger.Info("Starting cluster agent")

	// Connect to hub
	if err := a.connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to hub: %w", err)
	}

	// Register with hub
	if err := a.register(ctx); err != nil {
		return fmt.Errorf("failed to register with hub: %w", err)
	}

	// Start heartbeat
	go a.heartbeat(ctx)

	// Start operation stream
	go a.streamOperations(ctx)

	a.logger.Info("Agent started successfully")

	// Start log streaming
	go a.streamLogs(ctx)

	// Start metrics streaming
	go a.streamMetrics(ctx)

	a.logger.Info("Agent started successfully")
	return nil
}

// Stop stops the agent
func (a *Agent) Stop() {
	a.logger.Info("Stopping cluster agent")
	close(a.stopCh)

	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			a.logger.Warn("failed to close connection", zap.Error(err))
		}
	}
}

// connect establishes connection to the hub
func (a *Agent) connect(ctx context.Context) error {
	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, // TODO: Configure proper TLS
	})

	// Set up connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Connect to hub
	conn, err := grpc.DialContext(ctx, a.config.HubURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial hub: %w", err)
	}

	a.conn = conn
	a.client = agentv1.NewAgentServiceClient(conn)
	return nil
}

// register registers the agent with the hub
func (a *Agent) register(ctx context.Context) error {
	// Get cluster information
	clusterInfo, err := a.kubeClient.GetClusterInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Generate cluster name from kubeconfig context or use default
	clusterName := a.getClusterName()

	// Create registration request with cluster name
	req := &agentv1.RegisterRequest{
		ClusterName:  clusterName,
		AgentVersion: "1.0.0",
		Fingerprint:  "agent-fingerprint", // TODO: Generate proper fingerprint
		ClusterInfo: &agentv1.ClusterInfo{
			KubernetesVersion: clusterInfo.KubernetesVersion,
			Platform:          clusterInfo.Platform,
			NodeCount:         int32(clusterInfo.NodeCount),
			Region:            "unknown", // TODO: Detect region
			Labels:            clusterInfo.Labels,
		},
	}

	// Send registration request
	resp, err := a.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Set the cluster ID assigned by the hub
	a.clusterID = resp.ClusterId
	a.sessionToken = resp.SessionToken

	a.logger.Info("Agent registered successfully",
		zap.String("cluster_id", a.clusterID),
		zap.String("session_token", a.sessionToken),
		zap.Int64("heartbeat_interval", resp.HeartbeatInterval),
	)

	return nil
}

// heartbeat sends periodic heartbeats to the hub
func (a *Agent) heartbeat(ctx context.Context) {
	a.logger.Debug("Starting heartbeat",
		zap.Duration("heartbeat_interval", a.config.HeartbeatInterval),
		zap.String("hub_url", a.config.HubURL))

	// Use a default heartbeat if not configured
	heartbeatInterval := a.config.HeartbeatInterval
	if heartbeatInterval <= 0 {
		heartbeatInterval = 30 * time.Second
		a.logger.Warn("Heartbeat interval not configured, using default",
			zap.Duration("default_interval", heartbeatInterval))
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.sendHeartbeat(ctx); err != nil {
				a.logger.Error("Failed to send heartbeat", zap.Error(err))
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to the hub
func (a *Agent) sendHeartbeat(ctx context.Context) error {
	// Get cluster status
	status, err := a.getClusterStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	req := &agentv1.HeartbeatRequest{
		ClusterId:    a.clusterID,
		SessionToken: a.sessionToken,
		Status:       status,
	}

	resp, err := a.client.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("heartbeat failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("heartbeat failed: %s", resp.Message)
	}

	return nil
}

// getClusterStatus gets the current cluster status
func (a *Agent) getClusterStatus(ctx context.Context) (*agentv1.ClusterStatus, error) {
	// Check cluster health
	if err := a.kubeClient.HealthCheck(ctx); err != nil {
		return &agentv1.ClusterStatus{
			Status:    "unhealthy",
			LastCheck: timestamppb.New(time.Now()),
			Issues:    []string{err.Error()},
		}, nil
	}

	// Get cluster info for node counts
	info, err := a.kubeClient.GetClusterInfo(ctx)
	if err != nil {
		return &agentv1.ClusterStatus{
			Status:    "degraded",
			LastCheck: timestamppb.New(time.Now()),
			Issues:    []string{fmt.Sprintf("failed to get cluster info: %v", err)},
		}, nil
	}

	status := "healthy"
	if info.ReadyNodes < info.NodeCount {
		status = "degraded"
	}

	return &agentv1.ClusterStatus{
		Status:     status,
		ReadyNodes: int32(info.ReadyNodes),
		TotalNodes: int32(info.NodeCount),
		LastCheck:  timestamppb.New(time.Now()),
	}, nil
}

// streamOperations streams operations from the hub
func (a *Agent) streamOperations(ctx context.Context) {
	a.logger.Debug("Starting operation stream")

	req := &agentv1.StreamOperationsRequest{
		ClusterId:    a.clusterID,
		SessionToken: a.sessionToken,
	}

	stream, err := a.client.StreamOperations(ctx, req)
	if err != nil {
		a.logger.Error("Failed to start operation stream", zap.Error(err))
		return
	}

	a.logger.Debug("Operation stream started successfully")

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		default:
			operation, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					a.logger.Info("Operation stream closed")
					return
				}
				a.logger.Error("Failed to receive operation", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			// Only process operation if we successfully received one
			if operation != nil {
				// Process operation in a goroutine
				go a.processOperation(ctx, operation)
			}
		}
	}
}

// processOperation processes a single operation
func (a *Agent) processOperation(ctx context.Context, operation *agentv1.Operation) {
	if operation == nil {
		a.logger.Error("Received nil operation, skipping")
		return
	}

	a.logger.Info("Processing operation",
		zap.String("operation_id", operation.Id),
		zap.String("type", operation.Type),
	)

	// Create cancellable context for this operation
	opCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Store cancel function for potential cancellation
	a.cancelOps[operation.Id] = cancel
	defer func() {
		delete(a.cancelOps, operation.Id)
	}()

	// Set operation as started
	// TODO: Report operation started

	var result *anypb.Any
	var success bool
	var message string

	// Process based on operation type with cancellation support
	done := make(chan struct{})
	go func() {
		defer close(done)
		switch operation.Type {
		case "apply":
			result, success, message = a.processApplyOperation(opCtx, operation)
		case "exec":
			result, success, message = a.processExecOperation(opCtx, operation)
		case "sync":
			result, success, message = a.processSyncOperation(opCtx, operation)
		default:
			success = false
			message = fmt.Sprintf("unknown operation type: %s", operation.Type)
		}
	}()

	// Wait for operation completion or cancellation
	select {
	case <-opCtx.Done():
		// Operation was cancelled
		success = false
		message = "Operation was cancelled"
		result = &anypb.Any{} // Empty result for cancelled operations
		a.logger.Info("Operation cancelled",
			zap.String("operation_id", operation.Id),
		)
	case <-done:
		// Operation completed normally
		a.logger.Info("Operation completed",
			zap.String("operation_id", operation.Id),
			zap.Bool("success", success),
			zap.String("message", message),
		)
	}

	// Report result
	if err := a.reportResult(ctx, operation.Id, success, message, result); err != nil {
		a.logger.Error("Failed to report operation result", zap.Error(err))
	}
}

// processApplyOperation processes an apply operation
func (a *Agent) processApplyOperation(ctx context.Context, operation *agentv1.Operation) (*anypb.Any, bool, string) {
	// TODO: Extract manifest from operation payload
	// TODO: Apply manifest using kubeClient
	return nil, false, "not implemented"
}

// processExecOperation processes an exec operation
func (a *Agent) processExecOperation(ctx context.Context, operation *agentv1.Operation) (*anypb.Any, bool, string) {
	// TODO: Extract command from operation payload
	// TODO: Execute command using kubeClient
	return nil, false, "not implemented"
}

// processSyncOperation processes a sync operation
func (a *Agent) processSyncOperation(ctx context.Context, operation *agentv1.Operation) (*anypb.Any, bool, string) {
	// TODO: Sync cluster resources
	return nil, false, "not implemented"
}

// reportResult reports the result of an operation
func (a *Agent) reportResult(ctx context.Context, operationID string, success bool, message string, result *anypb.Any) error {
	req := &agentv1.ReportResultRequest{
		OperationId: operationID,
		ClusterId:   a.clusterID,
		Success:     success,
		Message:     message,
		Result:      result,
		CompletedAt: timestamppb.New(time.Now()),
	}

	resp, err := a.client.ReportResult(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to report result: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("result reporting failed: %s", resp.Message)
	}

	return nil
}

// streamLogs streams logs to the hub
func (a *Agent) streamLogs(ctx context.Context) {
	// TODO: Implement log streaming
	a.logger.Info("Log streaming not implemented yet")
}

// streamMetrics streams metrics to the hub
func (a *Agent) streamMetrics(ctx context.Context) {
	// TODO: Implement metrics streaming
	a.logger.Info("Metrics streaming not implemented yet")
}

// CancelOperation cancels a running operation
func (a *Agent) CancelOperation(operationID string) error {
	a.logger.Info("Cancellation requested",
		zap.String("operation_id", operationID),
	)

	if cancel, exists := a.cancelOps[operationID]; exists {
		cancel()
		a.logger.Info("Operation cancelled",
			zap.String("operation_id", operationID),
		)
		return nil
	}

	return fmt.Errorf("operation not found or not running: %s", operationID)
}

// getClusterName generates a meaningful cluster name
func (a *Agent) getClusterName() string {
	// Try to get cluster name from environment variable
	if name := os.Getenv("MCKMA_CLUSTER_NAME"); name != "" {
		return name
	}

	// Try to get from kubeconfig context
	if a.kubeClient != nil {
		// TODO: Extract context name from kubeconfig
		// For now, use a default name with timestamp
		return fmt.Sprintf("cluster-%d", time.Now().Unix())
	}

	// Fallback to hostname
	if hostname, err := os.Hostname(); err == nil {
		return fmt.Sprintf("cluster-%s", hostname)
	}

	// Final fallback
	return fmt.Sprintf("cluster-%d", time.Now().Unix())
}
