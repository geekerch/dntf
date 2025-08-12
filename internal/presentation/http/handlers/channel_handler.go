package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"notification/internal/application/channel/dtos"
	"notification/internal/application/channel/usecases"
)

// ChannelHandler handles HTTP requests for channel operations
type ChannelHandler struct {
	createUseCase *usecases.CreateChannelUseCase
	getUseCase    *usecases.GetChannelUseCase
	listUseCase   *usecases.ListChannelsUseCase
	updateUseCase *usecases.UpdateChannelUseCase
	deleteUseCase *usecases.DeleteChannelUseCase
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(
	createUseCase *usecases.CreateChannelUseCase,
	getUseCase *usecases.GetChannelUseCase,
	listUseCase *usecases.ListChannelsUseCase,
	updateUseCase *usecases.UpdateChannelUseCase,
	deleteUseCase *usecases.DeleteChannelUseCase,
) *ChannelHandler {
	return &ChannelHandler{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		listUseCase:   listUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// CreateChannel handles POST /api/v1/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var request dtos.CreateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.createUseCase.Execute(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create channel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetChannel handles GET /api/v1/channels/:id
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	response, err := h.getUseCase.Execute(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Channel not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListChannels handles GET /api/v1/channels
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	var request dtos.ListChannelsRequest

	// Parse query parameters
	if channelType := c.Query("channelType"); channelType != "" {
		request.ChannelType = channelType
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		request.Tags = tags
	}

	if skipCount := c.Query("skipCount"); skipCount != "" {
		if count, err := strconv.Atoi(skipCount); err == nil {
			request.SkipCount = count
		}
	}

	if maxResultCount := c.Query("maxResultCount"); maxResultCount != "" {
		if count, err := strconv.Atoi(maxResultCount); err == nil {
			request.MaxResultCount = count
		}
	}

	// Set default values
	if request.MaxResultCount <= 0 {
		request.MaxResultCount = 20
	}

	response, err := h.listUseCase.Execute(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list channels",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateChannel handles PUT /api/v1/channels/:id
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	var request dtos.UpdateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set the channel ID from URL parameter
	request.ChannelID = channelID

	response, err := h.updateUseCase.Execute(c.Request.Context(), channelID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update channel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteChannel handles DELETE /api/v1/channels/:id
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Channel ID is required",
		})
		return
	}

	response, err := h.deleteUseCase.Execute(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete channel",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}