package cluster

import "errors"

// Domain errors for cluster operations
var (
	ErrClusterNotFound             = errors.New("cluster not found")
	ErrClusterAlreadyExists        = errors.New("cluster already exists")
	ErrInvalidClusterMode          = errors.New("invalid cluster mode")
	ErrClusterNotConnected         = errors.New("cluster not connected")
	ErrClusterNotHealthy           = errors.New("cluster not healthy")
	ErrInvalidClusterID            = errors.New("invalid cluster ID")
	ErrClusterNameRequired         = errors.New("cluster name is required")
	ErrClusterModeRequired         = errors.New("cluster mode is required")
	ErrClusterStatusInvalid        = errors.New("invalid cluster status")
	ErrClusterLabelsInvalid        = errors.New("invalid cluster labels")
	ErrClusterResourcesNotFound    = errors.New("cluster resources not found")
	ErrClusterResourcesUnavailable = errors.New("cluster resources unavailable")
)
