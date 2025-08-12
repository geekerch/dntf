package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"notification/internal/application/cqrs"
	channelcqrs "notification/internal/application/cqrs/channel"
	"notification/internal/application/channel/dtos"
	"notification/pkg/logger"
)

// CQRSChannelHandler handles HTTP requests for channel operations using CQRS
type CQRSChannelHandler struct {
	cqrsFacade *cqrs.CQRSFacade
}

// NewCQRSChannelHandler creates a new CQRS channel handler
func NewCQRSChannelHandler(cqrsFacade *cqrs.CQRSFacade) *CQRSChannelHandler {
	return &CQRSChannelHandler{
		cqrsFacade: cqrsFacade,
	}
}

// CreateChannel handles POST /api/v1/channels using CQRS
func (h *CQRSChannelHandler) CreateChannel(c *gin.Context) {
	var request dtos.CreateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Create command
	command := channelcqrs.NewCreateChannelCommand(&request)
	
	// Set user context if available
	if userID, exists := c.Get("user_id"); exists {
		command.UserID = userID.(string)
	}
	if traceID, exists := c.Get("request_id"); exists {
		command.TraceID = traceID.(string)
	}

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), command)
	if err != nil {
		logger.Error("Failed to execute create channel command",
			zap.String("command_id", command.GetCommandID()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create channel",
			"details": err.Error(),
		})
		return
	}

	if !result.Success {
		logger.Error("Create channel command failed",
			zap.String("command_id", command.GetCommandID()),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create channel",
			"details": result.Error.Error(),
		})
		return
	}

	logger.Info("Channel created successfully",
		zap.String("command_id", command.GetCommandID()),
		zap.Duration("duration", result.Duration))

	c.Header("X-Command-ID", result.CommandID)
	c.JSON(http.StatusCreated, result.Data)
}

// GetChannel handles GET /api/v1/channels/:id using CQRS
func (h *CQRSChannelHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	// Create query
	query := channelcqrs.NewGetChannelQuery(channelID)
	
	// Set user context if available
	if userID, exists := c.Get("user_id"); exists {
		query.UserID = userID.(string)
	}
	if traceID, exists := c.Get("request_id"); exists {
		query.TraceID = traceID.(string)
	}

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to execute get channel query",
			zap.String("query_id", query.GetQueryID()),
			zap.String("channel_id", channelID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Channel not found",
			"details": err.Error(),
		})
		return
	}

	if !result.Success {
		logger.Error("Get channel query failed",
			zap.String("query_id", query.GetQueryID()),
			zap.Error(result.Error))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Channel not found",
			"details": result.Error.Error(),
		})
		return
	}

	c.Header("X-Query-ID", result.QueryID)
	if result.CacheHit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	
	c.JSON(http.StatusOK, result.Data)
}

// ListChannels handles GET /api/v1/channels using CQRS
func (h *CQRSChannelHandler) ListChannels(c *gin.Context) {
	// Create query
	query := channelcqrs.NewListChannelsQuery()

	// Parse query parameters
	if channelType := c.Query("channelType"); channelType != "" {
		query.WithChannelType(channelType)
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		query.WithTags(tags)
	}

	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			query.WithEnabled(enabled)
		}
	}

	// Parse pagination
	offset := 0
	limit := 20
	if skipCount := c.Query("skipCount"); skipCount != "" {
		if count, err := strconv.Atoi(skipCount); err == nil {
			offset = count
		}
	}
	if maxResultCount := c.Query("maxResultCount"); maxResultCount != "" {
		if count, err := strconv.Atoi(maxResultCount); err == nil && count > 0 {
			limit = count
		}
	}
	query.WithPagination(offset, limit)

	// Parse sorting
	if sortField := c.Query("sortField"); sortField != "" {
		sortOrder := c.DefaultQuery("sortOrder", "asc")
		query.WithSorting(sortField, sortOrder)
	}

	// Set user context if available
	if userID, exists := c.Get("user_id"); exists {
		query.UserID = userID.(string)
	}
	if traceID, exists := c.Get("request_id"); exists {
		query.TraceID = traceID.(string)
	}

	// Execute query
	result, err := h.cqrsFacade.Query(c.Request.Context(), query)
	if err != nil {
		logger.Error("Failed to execute list channels query",
			zap.String("query_id", query.GetQueryID()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list channels",
			"details": err.Error(),
		})
		return
	}

	if !result.Success {
		logger.Error("List channels query failed",
			zap.String("query_id", query.GetQueryID()),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list channels",
			"details": result.Error.Error(),
		})
		return
	}

	c.Header("X-Query-ID", result.QueryID)
	if result.CacheHit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	c.JSON(http.StatusOK, result.Data)
}

// UpdateChannel handles PUT /api/v1/channels/:id using CQRS
func (h *CQRSChannelHandler) UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	var request dtos.UpdateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Create command
	command := channelcqrs.NewUpdateChannelCommand(channelID, &request)
	
	// Set user context if available
	if userID, exists := c.Get("user_id"); exists {
		command.UserID = userID.(string)
	}
	if traceID, exists := c.Get("request_id"); exists {
		command.TraceID = traceID.(string)
	}

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), command)
	if err != nil {
		logger.Error("Failed to execute update channel command",
			zap.String("command_id", command.GetCommandID()),
			zap.String("channel_id", channelID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update channel",
			"details": err.Error(),
		})
		return
	}

	if !result.Success {
		logger.Error("Update channel command failed",
			zap.String("command_id", command.GetCommandID()),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update channel",
			"details": result.Error.Error(),
		})
		return
	}

	logger.Info("Channel updated successfully",
		zap.String("command_id", command.GetCommandID()),
		zap.String("channel_id", channelID),
		zap.Duration("duration", result.Duration))

	c.Header("X-Command-ID", result.CommandID)
	c.JSON(http.StatusOK, result.Data)
}

// DeleteChannel handles DELETE /api/v1/channels/:id using CQRS
func (h *CQRSChannelHandler) DeleteChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	// Create command
	command := channelcqrs.NewDeleteChannelCommand(channelID)
	
	// Set user context if available
	if userID, exists := c.Get("user_id"); exists {
		command.UserID = userID.(string)
	}
	if traceID, exists := c.Get("request_id"); exists {
		command.TraceID = traceID.(string)
	}

	// Execute command
	result, err := h.cqrsFacade.Send(c.Request.Context(), command)
	if err != nil {
		logger.Error("Failed to execute delete channel command",
			zap.String("command_id", command.GetCommandID()),
			zap.String("channel_id", channelID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete channel",
			"details": err.Error(),
		})
		return
	}

	if !result.Success {
		logger.Error("Delete channel command failed",
			zap.String("command_id", command.GetCommandID()),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete channel",
			"details": result.Error.Error(),
		})
		return
	}

	logger.Info("Channel deleted successfully",
		zap.String("command_id", command.GetCommandID()),
		zap.String("channel_id", channelID),
		zap.Duration("duration", result.Duration))

	c.Header("X-Command-ID", result.CommandID)
	c.JSON(http.StatusOK, result.Data)
}