package grpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
	agentv1 "github.com/rizesky/mckmt/api/proto/agent/v1"
	"github.com/rizesky/mckmt/internal/metrics"
	"github.com/rizesky/mckmt/internal/repo"
)

// Server represents the gRPC server for agent communication
type Server struct {
	agentv1.UnimplementedAgentServiceServer
	clusters   repo.ClusterRepository
	operations repo.OperationRepository
	metrics    *metrics.Metrics
	logger     *zap.Logger
	agents     map[string]*AgentConnection // cluster_id -> connection
}

// AgentConnection represents a connected agent
type AgentConnection struct {
	ClusterID     string
	AgentVersion  string
	LastHeartbeat time.Time
	Stream        chan *Operation
}

// Operation represents a task for the agent
type Operation struct {
	ID        string                 `json:"id"`
	ClusterID string                 `json:"cluster_id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Timeout   int32                  `json:"timeout_seconds"`
}

// NewServer creates a new gRPC server
func NewServer(clusters repo.ClusterRepository, operations repo.OperationRepository, metrics *metrics.Metrics, logger *zap.Logger) *Server {
	return &Server{
		clusters:   clusters,
		operations: operations,
		metrics:    metrics,
		logger:     logger,
		agents:     make(map[string]*AgentConnection),
	}
}

// Register handles agent registration
func (s *Server) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
	s.logger.Info("Agent registration request",
		zap.String("cluster_name", req.ClusterName),
		zap.String("agent_version", req.AgentVersion),
	)

	// Validate cluster name
	if req.ClusterName == "" {
		s.logger.Error("Cluster name is required")
		return &agentv1.RegisterResponse{
			Success: false,
			Message: "Cluster name is required",
		}, status.Error(codes.InvalidArgument, "Cluster name is required")
	}

	// Check if cluster with this name already exists
	var clusterID uuid.UUID
	existingCluster, err := s.clusters.GetByName(ctx, req.ClusterName)
	if err == nil {
		// Cluster exists, use its ID
		clusterID = existingCluster.ID
		s.logger.Info("Found existing cluster",
			zap.String("cluster_name", req.ClusterName),
			zap.String("cluster_id", clusterID.String()),
		)
	} else {
		// Cluster doesn't exist, generate new ID
		clusterID = uuid.New()
		s.logger.Info("Creating new cluster",
			zap.String("cluster_name", req.ClusterName),
			zap.String("cluster_id", clusterID.String()),
		)
	}

	// Check if cluster exists, if not create it
	cluster, err := s.clusters.GetByID(ctx, clusterID)
	if err != nil {
		// Cluster doesn't exist, create it
		s.logger.Info("Creating new cluster", zap.String("cluster_id", clusterID.String()))

		// Convert ClusterInfo to labels
		labels := make(repo.Labels)
		if req.ClusterInfo != nil {
			labels["kubernetes_version"] = req.ClusterInfo.KubernetesVersion
			labels["platform"] = req.ClusterInfo.Platform
			labels["node_count"] = fmt.Sprintf("%d", req.ClusterInfo.NodeCount)
			labels["region"] = req.ClusterInfo.Region

			// Add custom labels from ClusterInfo
			for k, v := range req.ClusterInfo.Labels {
				labels[k] = v
			}
		}

		// Create new cluster
		cluster = &repo.Cluster{
			ID:          clusterID,
			Name:        req.ClusterName, // Use provided cluster name
			Description: fmt.Sprintf("Cluster managed by agent %s", req.AgentVersion),
			Labels:      labels,
			Status:      "connected",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.clusters.Create(ctx, cluster); err != nil {
			s.logger.Error("Failed to create cluster", zap.Error(err))
			return &agentv1.RegisterResponse{
				Success: false,
				Message: "Failed to create cluster",
			}, status.Error(codes.Internal, "Failed to create cluster")
		}
	} else {
		// Cluster exists, update it with new info
		s.logger.Info("Updating existing cluster", zap.String("cluster_id", clusterID.String()))

		// Update cluster info if provided
		if req.ClusterInfo != nil {
			// Update labels with new cluster info
			if cluster.Labels == nil {
				cluster.Labels = make(repo.Labels)
			}
			cluster.Labels["kubernetes_version"] = req.ClusterInfo.KubernetesVersion
			cluster.Labels["platform"] = req.ClusterInfo.Platform
			cluster.Labels["node_count"] = fmt.Sprintf("%d", req.ClusterInfo.NodeCount)
			cluster.Labels["region"] = req.ClusterInfo.Region

			// Add custom labels from ClusterInfo
			for k, v := range req.ClusterInfo.Labels {
				cluster.Labels[k] = v
			}
		}

		// Update cluster status and timestamp
		cluster.Status = "connected"
		cluster.UpdatedAt = time.Now()

		if err := s.clusters.Update(ctx, cluster); err != nil {
			s.logger.Error("Failed to update cluster", zap.Error(err))
			return &agentv1.RegisterResponse{
				Success: false,
				Message: "Failed to update cluster",
			}, status.Error(codes.Internal, "Failed to update cluster")
		}
	}

	// Create agent connection
	connection := &AgentConnection{
		ClusterID:     clusterID.String(),
		AgentVersion:  req.AgentVersion,
		LastHeartbeat: time.Now(),
		Stream:        make(chan *Operation, 100),
	}
	s.agents[clusterID.String()] = connection

	// Update metrics
	s.metrics.SetAgentsConnected(clusterID.String(), req.AgentVersion, 1)

	s.logger.Info("Agent registered successfully",
		zap.String("cluster_name", req.ClusterName),
		zap.String("cluster_id", clusterID.String()),
		zap.String("agent_version", req.AgentVersion),
	)

	return &agentv1.RegisterResponse{
		Success:           true,
		Message:           "Registration successful",
		ClusterId:         cluster.ID.String(),
		SessionToken:      "session-" + cluster.ID.String(),
		HeartbeatInterval: 30,
	}, nil
}

// Heartbeat handles agent heartbeats
func (s *Server) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	connection, exists := s.agents[req.ClusterId]
	if !exists {
		return &agentv1.HeartbeatResponse{
			Success: false,
			Message: "Agent not registered",
		}, status.Error(codes.NotFound, "Agent not registered")
	}

	// Update heartbeat
	connection.LastHeartbeat = time.Now()

	// Update cluster status
	clusterStatus := "connected"
	if req.Status != nil {
		clusterStatus = req.Status.Status
	}

	clusterID, err := uuid.Parse(req.ClusterId)
	if err != nil {
		s.logger.Error("Invalid cluster ID", zap.String("cluster_id", req.ClusterId))
		return &agentv1.HeartbeatResponse{
			Success: false,
			Message: "Invalid cluster ID",
		}, status.Error(codes.InvalidArgument, "Invalid cluster ID")
	}

	if err := s.clusters.UpdateLastSeen(ctx, clusterID); err != nil {
		s.logger.Error("Failed to update cluster last seen", zap.Error(err))
	}

	// Update metrics
	s.metrics.RecordAgentHeartbeat(req.ClusterId, clusterStatus)
	s.metrics.SetAgentLastHeartbeat(req.ClusterId, float64(time.Now().Unix()))

	return &agentv1.HeartbeatResponse{
		Success: true,
		Message: "Heartbeat received",
	}, nil
}

// StreamOperations streams operations to agents
func (s *Server) StreamOperations(req *agentv1.StreamOperationsRequest, stream grpc.ServerStreamingServer[agentv1.Operation]) error {
	connection, exists := s.agents[req.ClusterId]
	if !exists {
		return status.Error(codes.NotFound, "Agent not registered")
	}

	s.logger.Info("Starting operation stream",
		zap.String("cluster_id", req.ClusterId),
	)

	// Send operations from the agent's stream
	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("Operation stream closed",
				zap.String("cluster_id", req.ClusterId),
			)
			return nil

		case operation := <-connection.Stream:
			// Convert custom Operation to generated Operation
			protoOp := &agentv1.Operation{
				Id:             operation.ID,
				ClusterId:      operation.ClusterID,
				Type:           operation.Type,
				TimeoutSeconds: operation.Timeout,
			}

			// Convert payload if present
			if operation.Payload != nil {
				// For now, we'll skip payload conversion as it requires more complex handling
				// TODO: Implement proper payload conversion
			}

			// Convert timestamp
			if !operation.CreatedAt.IsZero() {
				protoOp.CreatedAt = timestamppb.New(operation.CreatedAt)
			}

			if err := stream.Send(protoOp); err != nil {
				s.logger.Error("Failed to send operation",
					zap.Error(err),
					zap.String("cluster_id", req.ClusterId),
					zap.String("operation_id", operation.ID),
				)
				return err
			}

			s.logger.Debug("Operation sent to agent",
				zap.String("cluster_id", req.ClusterId),
				zap.String("operation_id", operation.ID),
				zap.String("type", operation.Type),
			)
		}
	}
}

// ReportResult handles operation result reporting
func (s *Server) ReportResult(ctx context.Context, req *agentv1.ReportResultRequest) (*agentv1.ReportResultResponse, error) {
	s.logger.Info("Operation result reported",
		zap.String("operation_id", req.OperationId),
		zap.String("cluster_id", req.ClusterId),
		zap.Bool("success", req.Success),
	)

	// Update operation in database
	operationID, err := uuid.Parse(req.OperationId)
	if err != nil {
		s.logger.Error("Invalid operation ID", zap.String("operation_id", req.OperationId))
		return &agentv1.ReportResultResponse{
			Success: false,
			Message: "Invalid operation ID",
		}, status.Error(codes.InvalidArgument, "Invalid operation ID")
	}

	operation, err := s.operations.GetByID(ctx, operationID)
	if err != nil {
		s.logger.Error("Operation not found", zap.Error(err))
		return &agentv1.ReportResultResponse{
			Success: false,
			Message: "Operation not found",
		}, status.Error(codes.NotFound, "Operation not found")
	}

	// Update operation status
	operationStatus := "success"
	if !req.Success {
		operationStatus = "failed"
	}

	if err := s.operations.UpdateStatus(ctx, operation.ID, operationStatus); err != nil {
		s.logger.Error("Failed to update operation status", zap.Error(err))
	}

	// Update operation result
	result := repo.Payload{
		"success":   req.Success,
		"message":   req.Message,
		"completed": req.CompletedAt,
	}

	if err := s.operations.UpdateResult(ctx, operation.ID, result); err != nil {
		s.logger.Error("Failed to update operation result", zap.Error(err))
	}

	// Update metrics
	s.metrics.RecordOperation(req.ClusterId, operation.Type, operationStatus, 0)

	return &agentv1.ReportResultResponse{
		Success: true,
		Message: "Result recorded",
	}, nil
}

// StreamLogs handles log streaming from agents
func (s *Server) StreamLogs(stream grpc.ClientStreamingServer[agentv1.LogEntry, agentv1.LogStreamResponse]) error {
	for {
		logEntry, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			s.logger.Error("Failed to receive log entry", zap.Error(err))
			return err
		}

		s.logger.Info("Log received from agent",
			zap.String("level", logEntry.Level),
			zap.String("message", logEntry.Message),
			zap.String("source", logEntry.Source),
		)

		// TODO: Store logs in database or forward to logging system
	}
}

// StreamMetrics handles metrics streaming from agents
func (s *Server) StreamMetrics(stream grpc.ClientStreamingServer[agentv1.MetricEntry, agentv1.MetricStreamResponse]) error {
	for {
		metricEntry, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			s.logger.Error("Failed to receive metric entry", zap.Error(err))
			return err
		}

		s.logger.Debug("Metric received from agent",
			zap.String("name", metricEntry.Name),
			zap.Float64("value", metricEntry.Value),
		)

		// TODO: Store metrics in Prometheus or forward to metrics system
	}
}

// QueueOperation queues an operation for an agent
func (s *Server) QueueOperation(clusterID string, operation *Operation) error {
	connection, exists := s.agents[clusterID]
	if !exists {
		return fmt.Errorf("agent not connected: %s", clusterID)
	}

	select {
	case connection.Stream <- operation:
		s.logger.Info("Operation queued for agent",
			zap.String("cluster_id", clusterID),
			zap.String("operation_id", operation.ID),
		)
		return nil
	default:
		return fmt.Errorf("agent operation queue full: %s", clusterID)
	}
}

// GetConnectedAgents returns list of connected agents
func (s *Server) GetConnectedAgents() []string {
	var agents []string
	for clusterID := range s.agents {
		agents = append(agents, clusterID)
	}
	return agents
}

// CancelOperation handles operation cancellation requests
func (s *Server) CancelOperation(ctx context.Context, req *agentv1.CancelOperationRequest) (*agentv1.CancelOperationResponse, error) {
	s.logger.Info("Operation cancellation requested",
		zap.String("operation_id", req.OperationId),
		zap.String("cluster_id", req.ClusterId),
		zap.String("reason", req.Reason),
	)

	// Validate operation ID
	operationID, err := uuid.Parse(req.OperationId)
	if err != nil {
		s.logger.Error("Invalid operation ID", zap.String("operation_id", req.OperationId))
		return &agentv1.CancelOperationResponse{
			Success: false,
			Message: "Invalid operation ID",
		}, status.Error(codes.InvalidArgument, "Invalid operation ID")
	}

	// Get operation to check if it can be cancelled
	operation, err := s.operations.GetByID(ctx, operationID)
	if err != nil {
		s.logger.Error("Operation not found", zap.Error(err))
		return &agentv1.CancelOperationResponse{
			Success: false,
			Message: "Operation not found",
		}, status.Error(codes.NotFound, "Operation not found")
	}

	// Check if operation can be cancelled
	if operation.Status == string(repo.OperationStatusSuccess) ||
		operation.Status == string(repo.OperationStatusFailed) ||
		operation.Status == string(repo.OperationStatusCancelled) {
		return &agentv1.CancelOperationResponse{
			Success: false,
			Message: fmt.Sprintf("Operation cannot be cancelled, current status: %s", operation.Status),
		}, status.Error(codes.FailedPrecondition, "Operation cannot be cancelled")
	}

	// Update operation status to cancelled
	if err := s.operations.UpdateStatus(ctx, operation.ID, string(repo.OperationStatusCancelled)); err != nil {
		s.logger.Error("Failed to update operation status", zap.Error(err))
		return &agentv1.CancelOperationResponse{
			Success: false,
			Message: "Failed to cancel operation",
		}, status.Error(codes.Internal, "Failed to cancel operation")
	}

	// Update operation result with cancellation info
	result := repo.Payload{
		"status":  "cancelled",
		"message": "Operation cancelled via gRPC",
		"reason":  req.Reason,
	}

	if err := s.operations.UpdateResult(ctx, operation.ID, result); err != nil {
		s.logger.Error("Failed to update operation result", zap.Error(err))
	}

	s.logger.Info("Operation cancelled successfully",
		zap.String("operation_id", req.OperationId),
		zap.String("cluster_id", req.ClusterId),
	)

	return &agentv1.CancelOperationResponse{
		Success: true,
		Message: "Operation cancelled successfully",
	}, nil
}

// DisconnectAgent removes an agent connection
func (s *Server) DisconnectAgent(clusterID string) {
	if connection, exists := s.agents[clusterID]; exists {
		close(connection.Stream)
		delete(s.agents, clusterID)
		s.metrics.SetAgentsConnected(clusterID, connection.AgentVersion, 0)
		s.logger.Info("Agent disconnected", zap.String("cluster_id", clusterID))
	}
}
