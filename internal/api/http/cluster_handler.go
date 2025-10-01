package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/rizesky/mckmt/internal/repo"
)

// ClusterHandler handles cluster-related HTTP requests
type ClusterHandler struct {
	clusterService ClusterManager
	logger         *zap.Logger
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(clusterService ClusterManager, logger *zap.Logger) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
		logger:         logger,
	}
}

// Note: With Go 1.22 enhanced routing, this method is no longer needed
// as the routing is handled directly in the router using pattern matching.

// ListClusters handles listing clusters
// @Summary List all clusters
// @Description Get a list of all registered clusters
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ClusterDTO
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters [get]
func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		if limit <= 0 {
			WriteErrorResponse(w, http.StatusBadRequest, "Limit must be positive")
			return
		}
	}

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}
		if offset < 0 {
			WriteErrorResponse(w, http.StatusBadRequest, "Offset must be non-negative")
			return
		}
	}

	clusters, err := h.clusterService.ListClusters(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list clusters", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list clusters")
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"clusters":    clusters,
		"total_count": len(clusters),
		"limit":       limit,
		"offset":      offset,
	})
}

// GetCluster handles getting a single cluster
// @Summary Get cluster by ID
// @Description Get a specific cluster by its ID
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cluster ID"
// @Success 200 {object} ClusterDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters/{id} [get]
func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	cluster, err := h.clusterService.GetCluster(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get cluster", zap.Error(err))
		if err.Error() == "not found" {
			WriteErrorResponse(w, http.StatusNotFound, "Cluster not found")
		} else {
			WriteErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	WriteJSONResponse(w, http.StatusOK, cluster)
}

// UpdateCluster handles updating a cluster
// @Summary Update cluster
// @Description Update an existing cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cluster ID"
// @Param cluster body repo.Cluster true "Cluster data"
// @Success 200 {object} ClusterDTO
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters/{id} [put]
func (h *ClusterHandler) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
		Status      string            `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cluster := &repo.Cluster{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Labels:      repo.Labels(req.Labels),
		Status:      req.Status,
		UpdatedAt:   time.Now(),
	}

	if err := h.clusterService.UpdateCluster(r.Context(), cluster.ID, cluster.Name, cluster.Description, map[string]string(cluster.Labels)); err != nil {
		h.logger.Error("Failed to update cluster", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update cluster")
		return
	}

	WriteJSONResponse(w, http.StatusOK, cluster)
}

// DeleteCluster handles deleting a cluster
// @Summary Delete cluster
// @Description Delete a cluster by ID
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cluster ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters/{id} [delete]
func (h *ClusterHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	if err := h.clusterService.DeleteCluster(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete cluster", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete cluster")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListClusterResources handles listing cluster resources
// @Summary List cluster resources
// @Description Get a list of resources in a specific cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cluster ID"
// @Param namespace query string false "Namespace filter"
// @Param kind query string false "Resource kind filter"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters/{id}/resources [get]
func (h *ClusterHandler) ListClusterResources(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")

	resources, err := h.clusterService.GetClusterResources(r.Context(), id, kind, namespace)
	if err != nil {
		h.logger.Error("Failed to get cluster resources", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get cluster resources")
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"cluster_id":  id,
		"resources":   resources,
		"total_count": len(resources),
		"kind":        kind,
		"namespace":   namespace,
	})
}

// ApplyManifests handles applying Kubernetes manifests
// @Summary Apply manifests to cluster
// @Description Apply Kubernetes manifests to a specific cluster
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cluster ID"
// @Param manifests body map[string]interface{} true "Kubernetes manifests"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /clusters/{id}/manifests [post]
func (h *ClusterHandler) ApplyManifests(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path using Chi
	idStr := chi.URLParam(r, "id")

	clusterID, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid cluster ID")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get the manifests file
	file, _, err := r.FormFile("manifests")
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "No manifests file provided")
		return
	}
	defer file.Close()

	// Read the manifests content
	manifests, err := io.ReadAll(file)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to read manifests file")
		return
	}

	// Create operation
	operation := &repo.Operation{
		ID:        uuid.New(),
		ClusterID: clusterID,
		Type:      "apply",
		Status:    "queued",
		Payload: repo.Payload{
			"manifests": string(manifests),
			"source":    "http_api",
		},
	}

	// Create operation in database
	err = h.clusterService.CreateOperation(r.Context(), operation)
	if err != nil {
		h.logger.Error("Failed to create operation", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create operation")
		return
	}

	// Queue operation for processing
	err = h.clusterService.QueueOperation(r.Context(), operation)
	if err != nil {
		h.logger.Error("Failed to queue operation", zap.Error(err))
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to queue operation")
		return
	}

	response := map[string]interface{}{
		"operation_id": operation.ID.String(),
		"status":       operation.Status,
		"message":      "Manifests queued for application",
	}

	WriteJSONResponse(w, http.StatusAccepted, response)
}
