package dtos

import (
	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/channel"
)

// CreateChannelRequest is the DTO for creating a channel.
type CreateChannelRequest struct {
	ChannelName    string                 `json:"channelName" binding:"required"`
	Description    string                 `json:"description"`
	Enabled        bool                   `json:"enabled"`
	ChannelType    string                 `json:"channelType" binding:"required"`
	TemplateID     string                 `json:"templateId"`
	CommonSettings CommonSettingsDTO      `json:"commonSettings" binding:"required"`
	Config         map[string]interface{} `json:"config" binding:"required"`
	Recipients     []RecipientDTO         `json:"recipients"`
	Tags           []string               `json:"tags"`
}

// UpdateChannelRequest is the DTO for updating a channel.
type UpdateChannelRequest struct {
	ChannelID      string                 `json:"channelId,omitempty"`
	ChannelName    string                 `json:"channelName" binding:"required"`
	Description    string                 `json:"description"`
	Enabled        bool                   `json:"enabled"`
	ChannelType    string                 `json:"channelType" binding:"required"`
	TemplateID     string                 `json:"templateId"`
	CommonSettings CommonSettingsDTO      `json:"commonSettings" binding:"required"`
	Config         map[string]interface{} `json:"config" binding:"required"`
	Recipients     []RecipientDTO         `json:"recipients"`
	Tags           []string               `json:"tags"`
}

// ListChannelsRequest is the DTO for listing channels.
type ListChannelsRequest struct {
	ChannelType    string   `form:"channelType" json:"channelType"`
	Tags           []string `form:"tags" json:"tags"`
	SkipCount      int      `form:"skipCount" json:"skipCount"`
	MaxResultCount int      `form:"maxResultCount" json:"maxResultCount"`
}

// ChannelResponse is the DTO for a channel response.
type ChannelResponse struct {
	ChannelID      string            `json:"channelId"`
	ChannelName    string            `json:"channelName"`
	Description    string            `json:"description"`
	Enabled        bool              `json:"enabled"`
	ChannelType    string            `json:"channelType"`
	TemplateID     string            `json:"templateId,omitempty"`
	CommonSettings CommonSettingsDTO `json:"commonSettings"`
	Config         map[string]interface{} `json:"config"`
	Recipients     []RecipientDTO    `json:"recipients"`
	Tags           []string          `json:"tags"`
	CreatedAt      int64             `json:"createdAt"`
	UpdatedAt      int64             `json:"updatedAt"`
	LastUsed       *int64            `json:"lastUsed,omitempty"`
}

// ChannelSummaryResponse is the DTO for a channel summary response (for list queries).
type ChannelSummaryResponse struct {
	ChannelID   string   `json:"channelId"`
	ChannelName string   `json:"channelName"`
	ChannelType string   `json:"channelType"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}

// ListChannelsResponse is the DTO for a list of channels.
type ListChannelsResponse struct {
	Items          []ChannelSummaryResponse `json:"items"`
	SkipCount      int                      `json:"skipCount"`
	MaxResultCount int                      `json:"maxResultCount"`
	TotalCount     int                      `json:"totalCount"`
	HasMore        bool                     `json:"hasMore"`
}

// DeleteChannelResponse is the DTO for a delete channel response.
type DeleteChannelResponse struct {
	ChannelID string `json:"channelId"`
	Deleted   bool   `json:"deleted"`
	DeletedAt int64  `json:"deletedAt"`
}

// CommonSettingsDTO is the DTO for common settings.
type CommonSettingsDTO struct {
	Timeout       int `json:"timeout" binding:"required,min=1"`
	RetryAttempts int `json:"retryAttempts" binding:"min=0"`
	RetryDelay    int `json:"retryDelay" binding:"min=0"`
}

// ToCommonSettings converts to a domain object.
func (dto CommonSettingsDTO) ToCommonSettings() (*shared.CommonSettings, error) {
	return shared.NewCommonSettings(dto.Timeout, dto.RetryAttempts, dto.RetryDelay)
}

// FromCommonSettings creates a DTO from a domain object.
func FromCommonSettings(settings *shared.CommonSettings) CommonSettingsDTO {
	return CommonSettingsDTO{
		Timeout:       settings.Timeout,
		RetryAttempts: settings.RetryAttempts,
		RetryDelay:    settings.RetryDelay,
	}
}

// RecipientDTO is the DTO for a recipient.
type RecipientDTO struct {
	Name   string `json:"name" binding:"required"`
	Email  string `json:"email,omitempty"`
	Target string `json:"target,omitempty"`
	Type   string `json:"type" binding:"required"`
}

// ToRecipient converts to a domain object.
func (dto RecipientDTO) ToRecipient() (*channel.Recipient, error) {
	return channel.NewRecipient(dto.Name, dto.Email, dto.Target, dto.Type)
}

// FromRecipient creates a DTO from a domain object.
func FromRecipient(recipient *channel.Recipient) RecipientDTO {
	return RecipientDTO{
		Name:   recipient.Name,
		Email:  recipient.Email,
		Target: recipient.Target,
		Type:   recipient.Type,
	}
}

// ToRecipientsSlice converts to a slice of recipients.
func ToRecipientsSlice(dtos []RecipientDTO) ([]*channel.Recipient, error) {
	recipients := make([]*channel.Recipient, 0, len(dtos))
	for _, dto := range dtos {
		recipient, err := dto.ToRecipient()
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

// FromRecipientsSlice creates a slice of DTOs from a slice of recipients.
func FromRecipientsSlice(recipients []*channel.Recipient) []RecipientDTO {
	dtos := make([]RecipientDTO, 0, len(recipients))
	for _, recipient := range recipients {
		dtos = append(dtos, FromRecipient(recipient))
	}
	return dtos
}