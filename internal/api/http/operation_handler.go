package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/operation"
)

// OperationHandler handles operation-related HTTP requests
type OperationHandler struct {
	operationService *operation.Service
	logger           *zap.Logger
}

// NewOperationHandler creates a new operation handler
func NewOperationHandler(operationService *operation.Service, logger *zap.Logger) *OperationHandler {
	return &OperationHandler{
		operationService: operationService,
		logger:           logger,
	}
}

// GetOperation handles getting a single operation
// @Summary Get operation by ID
// @Description Get a specific operation by its ID
// @Tags operations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Operation ID"
// @Success 200 {object} OperationDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /operations/{id} [get]
func (h *OperationHandler) GetOperation(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid operation ID")
		return
	}

	operation, err := h.operationService.GetOperation(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get operation", zap.Error(err))
		WriteErrorResponse(w, http.StatusNotFound, "Operation not found")
		return
	}

	WriteJSONResponse(w, http.StatusOK, operation)
}

// ListOperationsByCluster handles listing operations for a cluster
// @Summary List operations by cluster
// @Description Get a list of operations for a specific cluster
// @Tags operations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "Cluster ID"
// @Param status query string false "Operation status filter"
// @Param limit query int false "Limit number of results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} OperationDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /operations/cluster/{clusterId} [get]
func (h *OperationHandler) ListOperationsByCluster(w http.ResponseWriter, r *http.Request) {
	// Extract cluster ID from URL path using Chi
	clusterIDStr := chi.URLParam(r, "clusterId")

	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	operations, err := h.operationService.ListOperationsByCluster(r.Context(), clusterID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list operations", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list operations")
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"operations":  operations,
		"total_count": len(operations),
		"cluster_id":  clusterID,
		"limit":       limit,
		"offset":      offset,
	})
}

// CancelOperation handles cancelling an operation
// @Summary Cancel operation
// @Description Cancel a running operation
// @Tags operations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Operation ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /operations/{id}/cancel [post]
func (h *OperationHandler) CancelOperation(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid operation ID")
		return
	}

	// Parse request body
	var req operation.CancelOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Cancel the operation
	err = h.operationService.CancelOperation(r.Context(), id, req.Reason)
	if err != nil {
		h.logger.Error("Failed to cancel operation", zap.Error(err))
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return success response
	response := operation.CancelOperationResponse{
		ID:      id,
		Status:  "cancelled",
		Message: "Operation cancelled successfully",
	}

	WriteJSONResponse(w, http.StatusOK, response)
}
