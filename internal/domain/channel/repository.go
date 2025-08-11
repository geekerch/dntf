package channel

import (
	"context"

	"channel-api/internal/domain/shared"
)

// ChannelRepository is the interface for the channel repository.
type ChannelRepository interface {
	// Save saves a channel.
	Save(ctx context.Context, channel *Channel) error
	
	// FindByID finds a channel by ID.
	FindByID(ctx context.Context, id *ChannelID) (*Channel, error)
	
	// FindByName finds a channel by name.
	FindByName(ctx context.Context, name *ChannelName) (*Channel, error)
	
	// FindAll finds all channels (supports pagination and filtering).
	FindAll(ctx context.Context, filter *ChannelFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*Channel], error)
	
	// Update updates a channel.
	Update(ctx context.Context, channel *Channel) error
	
	// Delete deletes a channel.
	Delete(ctx context.Context, id *ChannelID) error
	
	// Exists checks if a channel exists.
	Exists(ctx context.Context, id *ChannelID) (bool, error)
	
	// ExistsByName checks if a channel with the specified name exists.
	ExistsByName(ctx context.Context, name *ChannelName) (bool, error)
}

// ChannelFilter is the filter for channels.
type ChannelFilter struct {
	ChannelType *shared.ChannelType `json:"channelType,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Enabled     *bool               `json:"enabled,omitempty"`
}

// NewChannelFilter creates a channel filter.
func NewChannelFilter() *ChannelFilter {
	return &ChannelFilter{}
}

// WithChannelType sets the channel type filter.
func (f *ChannelFilter) WithChannelType(channelType shared.ChannelType) *ChannelFilter {
	f.ChannelType = &channelType
	return f
}

// WithTags sets the tag filter.
func (f *ChannelFilter) WithTags(tags []string) *ChannelFilter {
	f.Tags = tags
	return f
}

// WithEnabled sets the enabled status filter.
func (f *ChannelFilter) WithEnabled(enabled bool) *ChannelFilter {
	f.Enabled = &enabled
	return f
}

// HasChannelTypeFilter checks if there is a channel type filter.
func (f *ChannelFilter) HasChannelTypeFilter() bool {
	return f.ChannelType != nil
}

// HasTagsFilter checks if there is a tag filter.
func (f *ChannelFilter) HasTagsFilter() bool {
	return len(f.Tags) > 0
}

// HasEnabledFilter checks if there is an enabled status filter.
func (f *ChannelFilter) HasEnabledFilter() bool {
	return f.Enabled != nil
}