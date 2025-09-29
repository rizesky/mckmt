package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/app/operations"
	"github.com/rizesky/mckmt/internal/auth"
)

// OperationHandler handles operation-related HTTP requests
type OperationHandler struct {
	operationService *operations.OperationService
	logger           *zap.Logger
	authMiddleware   *auth.AuthMiddleware
}

// NewOperationHandler creates a new operation handler
func NewOperationHandler(operationService *operations.OperationService, logger *zap.Logger) *OperationHandler {
	// Create JWT manager with default settings
	jwtManager := auth.NewJWTManager("your-secret-key", 24*time.Hour)

	return &OperationHandler{
		operationService: operationService,
		logger:           logger,
		authMiddleware:   auth.NewAuthMiddleware(jwtManager, logger),
	}
}

// RegisterRoutes registers operation routes
func (h *OperationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/operations", func(r chi.Router) {
		r.Use(h.authMiddleware.RequireAuth)
		r.Get("/{id}", h.GetOperation)
		r.Get("/cluster/{clusterId}", h.ListOperationsByCluster)
		r.Post("/{id}/cancel", h.CancelOperation)
	})
}

// GetOperation handles getting a single operation
func (h *OperationHandler) GetOperation(w http.ResponseWriter, r *http.Request) {
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
func (h *OperationHandler) ListOperationsByCluster(w http.ResponseWriter, r *http.Request) {
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
func (h *OperationHandler) CancelOperation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid operation ID")
		return
	}

	// Parse request body
	var req operations.CancelOperationRequest
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
	response := operations.CancelOperationResponse{
		ID:      id,
		Status:  "cancelled",
		Message: "Operation cancelled successfully",
	}

	WriteJSONResponse(w, http.StatusOK, response)
}
