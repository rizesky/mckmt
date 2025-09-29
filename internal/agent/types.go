package agent

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Protobuf message types (simplified versions for now)

type RegisterRequest struct {
	ClusterName  string // Human-readable cluster name/identifier
	AgentVersion string
	Fingerprint  string
	ClusterInfo  *ClusterInfo
}

type RegisterResponse struct {
	Success           bool
	Message           string
	ClusterId         string
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

type Operation struct {
	Id             string
	ClusterId      string
	Type           string
	Payload        *anypb.Any
	CreatedAt      *timestamppb.Timestamp
	TimeoutSeconds int32
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
	Platform          string
	NodeCount         int32
	Region            string
	Labels            map[string]string
}

type ClusterStatus struct {
	Status     string
	ReadyNodes int32
	TotalNodes int32
	Issues     []string
	LastCheck  time.Time
}

// AgentServiceClient interface for gRPC client
type AgentServiceClient interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error)
	StreamOperations(ctx context.Context, req *StreamOperationsRequest) (AgentService_StreamOperationsClient, error)
	ReportResult(ctx context.Context, req *ReportResultRequest) (*ReportResultResponse, error)
	StreamLogs(ctx context.Context) (AgentService_StreamLogsClient, error)
	StreamMetrics(ctx context.Context) (AgentService_StreamMetricsClient, error)
}

// Stream clients
type AgentService_StreamOperationsClient interface {
	Recv() (*Operation, error)
	CloseSend() error
}

type AgentService_StreamLogsClient interface {
	Send(*LogEntry) error
	CloseAndRecv() (*LogStreamResponse, error)
}

type AgentService_StreamMetricsClient interface {
	Send(*MetricEntry) error
	CloseAndRecv() (*MetricStreamResponse, error)
}

// Mock implementation for now
type mockAgentServiceClient struct{}

func (m *mockAgentServiceClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	return &RegisterResponse{
		Success:           true,
		Message:           "Registration successful",
		SessionToken:      "mock-session-token",
		HeartbeatInterval: 30,
	}, nil
}

func (m *mockAgentServiceClient) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	return &HeartbeatResponse{
		Success: true,
		Message: "Heartbeat received",
	}, nil
}

func (m *mockAgentServiceClient) StreamOperations(ctx context.Context, req *StreamOperationsRequest) (AgentService_StreamOperationsClient, error) {
	return &mockStreamOperationsClient{}, nil
}

func (m *mockAgentServiceClient) ReportResult(ctx context.Context, req *ReportResultRequest) (*ReportResultResponse, error) {
	return &ReportResultResponse{
		Success: true,
		Message: "Result reported",
	}, nil
}

func (m *mockAgentServiceClient) StreamLogs(ctx context.Context) (AgentService_StreamLogsClient, error) {
	return &mockStreamLogsClient{}, nil
}

func (m *mockAgentServiceClient) StreamMetrics(ctx context.Context) (AgentService_StreamMetricsClient, error) {
	return &mockStreamMetricsClient{}, nil
}

type mockStreamOperationsClient struct{}

func (m *mockStreamOperationsClient) Recv() (*Operation, error) {
	// Simulate no operations for now
	time.Sleep(1 * time.Second)
	return nil, nil
}

func (m *mockStreamOperationsClient) CloseSend() error {
	return nil
}

type mockStreamLogsClient struct{}

func (m *mockStreamLogsClient) Send(*LogEntry) error {
	return nil
}

func (m *mockStreamLogsClient) CloseAndRecv() (*LogStreamResponse, error) {
	return &LogStreamResponse{Success: true}, nil
}

type mockStreamMetricsClient struct{}

func (m *mockStreamMetricsClient) Send(*MetricEntry) error {
	return nil
}

func (m *mockStreamMetricsClient) CloseAndRecv() (*MetricStreamResponse, error) {
	return &MetricStreamResponse{Success: true}, nil
}

// NewAgentServiceClient creates a new agent service client
func NewAgentServiceClient(conn interface{}) AgentServiceClient {
	// For now, return mock implementation
	return &mockAgentServiceClient{}
}
