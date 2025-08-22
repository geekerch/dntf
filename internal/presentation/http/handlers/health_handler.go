package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "notification/internal/application/health/dtos" // Required for Swagger documentation
	"notification/internal/application/health/usecases"
)

// HealthHandler handles health check endpoints following Clean Architecture
type HealthHandler struct {
	getSystemHealthUseCase *usecases.GetSystemHealthUseCase
	getLivenessUseCase     *usecases.GetLivenessUseCase
	getLegacyHealthUseCase *usecases.GetLegacyHealthUseCase
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	getSystemHealthUseCase *usecases.GetSystemHealthUseCase,
	getLivenessUseCase *usecases.GetLivenessUseCase,
	getLegacyHealthUseCase *usecases.GetLegacyHealthUseCase,
) *HealthHandler {
	return &HealthHandler{
		getSystemHealthUseCase: getSystemHealthUseCase,
		getLivenessUseCase:     getLivenessUseCase,
		getLegacyHealthUseCase: getLegacyHealthUseCase,
	}
}

// Healthz provides minimal liveness check
// @Summary Minimal liveness check
// @Description Provides a lightweight, fast response used for liveness checks by Kubernetes, load balancers, or uptime monitoring tools
// @Tags health
// @Produce json
// @Success 200 {object} dtos.LivenessResponse
// @Failure 500 {object} map[string]string
// @Router /healthz [get]
func (h *HealthHandler) Healthz(c *gin.Context) {
	response, err := h.getLivenessUseCase.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform liveness check",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthStatus provides detailed health check
// @Summary Detailed health check
// @Description Provides comprehensive view of service health including dependencies
// @Tags health
// @Produce json
// @Success 200 {object} dtos.DetailedHealthResponse
// @Failure 503 {object} dtos.DetailedHealthResponse
// @Router /health-status [get]
func (h *HealthHandler) HealthStatus(c *gin.Context) {
	response, err := h.getSystemHealthUseCase.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform system health check",
		})
		return
	}

	statusCode := http.StatusOK
	if response.Status == "Unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Health provides legacy health check (for backward compatibility)
// @Summary Legacy health check
// @Description Legacy health endpoint for backward compatibility
// @Tags health
// @Produce json
// @Success 200 {object} dtos.LegacyHealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response, err := h.getLegacyHealthUseCase.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get health status",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
