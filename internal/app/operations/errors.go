package operations

import "errors"

// Domain errors for operation operations
var (
	ErrOperationNotFound        = errors.New("operation not found")
	ErrOperationAlreadyExists   = errors.New("operation already exists")
	ErrInvalidOperationType     = errors.New("invalid operation type")
	ErrInvalidOperationStatus   = errors.New("invalid operation status")
	ErrOperationNotRunning      = errors.New("operation is not running")
	ErrOperationAlreadyRunning  = errors.New("operation is already running")
	ErrOperationAlreadyFinished = errors.New("operation is already finished")
	ErrOperationCannotStart     = errors.New("operation cannot be started")
	ErrOperationCannotFinish    = errors.New("operation cannot be finished")
	ErrOperationCannotCancel    = errors.New("operation cannot be cancelled")
	ErrInvalidOperationID       = errors.New("invalid operation ID")
	ErrOperationTypeRequired    = errors.New("operation type is required")
	ErrOperationClusterRequired = errors.New("operation cluster ID is required")
	ErrOperationPayloadInvalid  = errors.New("invalid operation payload")
	ErrOperationResultInvalid   = errors.New("invalid operation result")
)
