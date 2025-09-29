package grpc

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// gRPC message types (simplified versions for now)

type RegisterRequest struct {
	ClusterId    string
	AgentVersion string
	Fingerprint  string
	ClusterInfo  *ClusterInfo
}

type RegisterResponse struct {
	Success           bool
	Message           string
	SessionToken      string
	HeartbeatInterval int64
}

type HeartbeatRequest struct {
	ClusterId    string
	SessionToken string
	Status       *ClusterStatus
}

type HeartbeatResponse struct {
	Success bool
	Message string
}

type StreamOperationsRequest struct {
	ClusterId    string
	SessionToken string
}

type ReportResultRequest struct {
	OperationId string
	ClusterId   string
	Success     bool
	Message     string
	Result      *anypb.Any
	CompletedAt time.Time
}

type ReportResultResponse struct {
	Success bool
	Message string
}

type LogEntry struct {
	Level     string
	Message   string
	Source    string
	Timestamp *timestamppb.Timestamp
	Fields    map[string]string
}

type LogStreamResponse struct {
	Success bool
	Message string
}

type MetricEntry struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp *timestamppb.Timestamp
}

type MetricStreamResponse struct {
	Success bool
	Message string
}

type ClusterInfo struct {
	KubernetesVersion string
	Platform         string
	NodeCount        int32
	Region           string
	Labels           map[string]string
}

type ClusterStatus struct {
	Status     string
	ReadyNodes int32
	TotalNodes int32
	Issues     []string
	LastCheck  time.Time
}

// AgentServiceServer interface for gRPC server
type AgentServiceServer interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error)
	StreamOperations(req *StreamOperationsRequest, stream AgentService_StreamOperationsServer) error
	ReportResult(ctx context.Context, req *ReportResultRequest) (*ReportResultResponse, error)
	StreamLogs(stream AgentService_StreamLogsServer) error
	StreamMetrics(stream AgentService_StreamMetricsServer) error
}

// Stream servers
type AgentService_StreamOperationsServer interface {
	Send(*Operation) error
	Context() context.Context
}

type AgentService_StreamLogsServer interface {
	SendAndClose(*LogStreamResponse) error
	Recv() (*LogEntry, error)
	Context() context.Context
}

type AgentService_StreamMetricsServer interface {
	SendAndClose(*MetricStreamResponse) error
	Recv() (*MetricEntry, error)
	Context() context.Context
}
