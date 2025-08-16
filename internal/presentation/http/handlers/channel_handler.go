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

// CreateChannel handles the creation of a new channel.
// @Summary      Create a new channel
// @Description  Creates a new channel with the provided details.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateChannelRequest true "Create Channel Request"
// @Success      201  {object}  map[string]interface{} "Success response with channel data"
// @Failure      400  {object}  map[string]interface{} "Bad Request - Invalid input or validation error"
// @Failure      409  {object}  map[string]interface{} "Conflict - Channel with the same name already exists"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Security     ApiKeyAuth
// @Router       /api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var request dtos.CreateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format: " + err.Error(),
			},
		})
		return
	}

	response, err := h.createUseCase.Execute(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "CREATE_CHANNEL_FAILED",
				"message": "Failed to create channel: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":  response,
		"error": nil,
	})
}

// GetChannel handles GET /api/v1/channels/:id
// @Summary      Get a channel by ID
// @Description  Retrieves a single channel's details using its unique identifier.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Channel ID"
// @Success      200  {object}  map[string]interface{} "Success response with channel data"
// @Failure      400  {object}  map[string]interface{} "Bad Request - Invalid channel ID format"
// @Failure      404  {object}  map[string]interface{} "Not Found - Channel with specified ID does not exist"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{id} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Channel ID is required",
			},
		})
		return
	}

	response, err := h.getUseCase.Execute(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "CHANNEL_NOT_FOUND",
				"message": "Channel not found: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// ListChannels handles GET /api/v1/channels
// @Summary      List all channels
// @Description  Retrieves a list of all channels, with optional filtering by channel type and tags, and pagination.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        channelType   query      string  false  "Filter by channel type (e.g., email, sms)"
// @Param        tags          query      []string  false  "Filter by tags (comma-separated)"  collectionFormat(csv)
// @Param        skipCount     query      int     false  "Number of records to skip for pagination"  default(0)
// @Param        maxResultCount query      int     false  "Maximum number of records to return per page (1-100)"  default(10)
// @Success      200  {object}  map[string]interface{} "Success response with channels list"
// @Failure      400  {object}  map[string]interface{} "Bad Request - Invalid query parameters"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels [get]
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
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "LIST_CHANNELS_FAILED",
				"message": "Failed to list channels: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// UpdateChannel handles PUT /api/v1/channels/:id
// @Summary      Update an existing channel
// @Description  Updates an existing channel's details using its unique identifier.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Channel ID"
// @Param        request body dtos.UpdateChannelRequest true "Update Channel Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{} "Bad Request - Invalid input or validation error"
// @Failure      404  {object}  map[string]interface{} "Not Found - Channel with specified ID does not exist"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{id} [put]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Channel ID is required",
			},
		})
		return
	}

	var request dtos.UpdateChannelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format: " + err.Error(),
			},
		})
		return
	}

	// Set the channel ID from URL parameter
	request.ChannelID = channelID

	response, err := h.updateUseCase.Execute(c.Request.Context(), channelID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "UPDATE_CHANNEL_FAILED",
				"message": "Failed to update channel: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}

// DeleteChannel handles DELETE /api/v1/channels/:id
// @Summary      Delete a channel by ID
// @Description  Deletes a channel using its unique identifier.
// @Tags         channels
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Channel ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{} "Bad Request - Invalid channel ID format"
// @Failure      404  {object}  map[string]interface{} "Not Found - Channel with specified ID does not exist"
// @Failure      500  {object}  map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Channel ID is required",
			},
		})
		return
	}

	response, err := h.deleteUseCase.Execute(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  nil,
			"error": map[string]interface{}{
				"code":    "DELETE_CHANNEL_FAILED",
				"message": "Failed to delete channel: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"error": nil,
	})
}