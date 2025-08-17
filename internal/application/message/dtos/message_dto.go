package dtos

import (
	"notification/internal/domain/message"
	"notification/internal/domain/shared"
)

// SendMessageRequest represents the request to send a message.
type SendMessageRequest struct {
	ChannelIDs       []string                  `json:"channelIds" validate:"required,min=1"`
	TemplateID       string                    `json:"templateId" validate:"required"`
	Recipients       []string                  `json:"recipients" validate:"required,min=1"`
	Variables        map[string]interface{}    `json:"variables,omitempty"`
	ChannelOverrides *message.ChannelOverrides `json:"channelOverrides,omitempty"`
	Settings         *shared.CommonSettings    `json:"settings,omitempty"`
}

// ListMessagesRequest represents the request to list messages.
type ListMessagesRequest struct {
	ChannelID      string `json:"channelId,omitempty"`
	Status         string `json:"status,omitempty"`
	SkipCount      int    `json:"skipCount,omitempty"`
	MaxResultCount int    `json:"maxResultCount,omitempty"`
}

// ListMessagesResponse represents the response for listing messages.
type ListMessagesResponse struct {
	Items          []*MessageResponse `json:"items"`
	SkipCount      int                `json:"skipCount"`
	MaxResultCount int                `json:"maxResultCount"`
	TotalCount     int                `json:"totalCount"`
	HasMore        bool               `json:"hasMore"`
}

// MessageResponse represents the response for a message.
type MessageResponse struct {
	ID               string                    `json:"id"`
	ChannelID        string                    `json:"channelId"`
	TemplateID       string                    `json:"templateId"`
	Recipients       []string                  `json:"recipients"`
	Variables        map[string]interface{}    `json:"variables,omitempty"`
	ChannelOverrides *message.ChannelOverrides `json:"channelOverrides,omitempty"`
	Status           message.MessageStatus     `json:"status"`
	Results          []*MessageResultResponse  `json:"results,omitempty"`
	Settings         *shared.CommonSettings    `json:"settings,omitempty"`
	CreatedAt        int64                     `json:"createdAt"`
}

// MessageResultResponse represents the response for a message result.
type MessageResultResponse struct {
	Recipient string                      `json:"recipient"`
	Status    message.MessageResultStatus `json:"status"`
	Error     string                      `json:"error,omitempty"`
	SentAt    *int64                      `json:"sentAt,omitempty"`
}

// ToMessageResponse converts a message entity to a response DTO.
func ToMessageResponse(m *message.Message) *MessageResponse {
	if m == nil {
		return nil
	}

	// Note: The current message entity structure doesn't match our DTO exactly
	// We'll need to adapt based on what's available
	response := &MessageResponse{
		ID:        m.ID().String(),
		Status:    m.Status(),
		CreatedAt: m.CreatedAt(),
	}

	// Get the first channel ID if available
	if m.ChannelIDs() != nil && m.ChannelIDs().Count() > 0 {
		channelIDs := m.ChannelIDs().ToSlice()
		if len(channelIDs) > 0 {
			response.ChannelID = channelIDs[0].String()
		}
	}

	if m.Variables() != nil {
		response.Variables = m.Variables().ToMap()
	}

	if m.ChannelOverrides() != nil {
		response.ChannelOverrides = m.ChannelOverrides()
	}

	// Convert results
	if len(m.Results()) > 0 {
		response.Results = make([]*MessageResultResponse, len(m.Results()))
		for i, result := range m.Results() {
			response.Results[i] = &MessageResultResponse{
				Status: result.Status(),
			}

			if result.Error() != nil {
				response.Results[i].Error = result.Error().Details
			}

			if result.SentAt() != nil {
				response.Results[i].SentAt = result.SentAt()
			}
		}
	}

	return response
}
