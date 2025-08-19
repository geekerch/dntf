package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"notification/internal/domain/channel"
	"notification/internal/domain/shared"
	"notification/internal/domain/template"
	"notification/internal/infrastructure/models"
)

// ChannelRepositoryImpl implements channel.ChannelRepository interface using GORM
type ChannelRepositoryImpl struct {
	db *gorm.DB
}

// NewChannelRepositoryImpl creates a new channel repository implementation
func NewChannelRepositoryImpl(db *gorm.DB) *ChannelRepositoryImpl {
	return &ChannelRepositoryImpl{
		db: db,
	}
}

// Save saves a channel to the database
func (r *ChannelRepositoryImpl) Save(ctx context.Context, ch *channel.Channel) error {
	model, err := r.toChannelModel(ch)
	if err != nil {
		return fmt.Errorf("failed to convert channel to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save channel: %w", err)
	}

	return nil
}

// FindByID finds a channel by its ID
func (r *ChannelRepositoryImpl) FindByID(ctx context.Context, id *channel.ChannelID) (*channel.Channel, error) {
	var model models.ChannelModel

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id.String()).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	return r.fromChannelModel(&model)
}

// FindByName finds a channel by its name
func (r *ChannelRepositoryImpl) FindByName(ctx context.Context, name *channel.ChannelName) (*channel.Channel, error) {
	var model models.ChannelModel

	err := r.db.WithContext(ctx).
		Where("name = ? AND deleted_at IS NULL", name.String()).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	return r.fromChannelModel(&model)
}

// FindAll finds all channels with filtering and pagination
func (r *ChannelRepositoryImpl) FindAll(ctx context.Context, filter *channel.ChannelFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*channel.Channel], error) {
	query := r.db.WithContext(ctx).Model(&models.ChannelModel{}).Where("deleted_at IS NULL")

	// Apply filters
	if filter.HasChannelTypeFilter() {
		query = query.Where("channel_type = ?", filter.ChannelType.String())
	}

	if filter.HasTagsFilter() {
		// For PostgreSQL, use array overlap operator
		if r.db.Dialector.Name() == "postgres" {
			query = query.Where("tags && ?", pq.StringArray(filter.Tags))
		} else {
			// For other databases, use JSON contains logic
			for _, tag := range filter.Tags {
				query = query.Where("JSON_EXTRACT(tags, '$') LIKE ?", "%"+tag+"%")
			}
		}
	}

	if filter.HasEnabledFilter() {
		query = query.Where("enabled = ?", *filter.Enabled)
	}

	// Count total records
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count channels: %w", err)
	}

	// Query channels with pagination
	var channelModels []models.ChannelModel
	err := query.
		Order("created_at DESC").
		Limit(pagination.MaxResultCount).
		Offset(pagination.SkipCount).
		Find(&channelModels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}

	// Convert to domain objects
	channels := make([]*channel.Channel, 0, len(channelModels))
	for _, model := range channelModels {
		ch, err := r.fromChannelModel(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to channel: %w", err)
		}
		channels = append(channels, ch)
	}

	// Calculate hasMore
	hasMore := pagination.SkipCount+len(channels) < int(totalCount)

	return &shared.PaginatedResult[*channel.Channel]{
		Items:          channels,
		SkipCount:      pagination.SkipCount,
		MaxResultCount: pagination.MaxResultCount,
		TotalCount:     int(totalCount),
		HasMore:        hasMore,
	}, nil
}

// Update updates a channel in the database
func (r *ChannelRepositoryImpl) Update(ctx context.Context, ch *channel.Channel) error {
	model, err := r.toChannelModel(ch)
	if err != nil {
		return fmt.Errorf("failed to convert channel to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// Delete deletes a channel from the database (hard delete)
func (r *ChannelRepositoryImpl) Delete(ctx context.Context, id *channel.ChannelID) error {
	if err := r.db.WithContext(ctx).Delete(&models.ChannelModel{}, "id = ?", id.String()).Error; err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// Exists checks if a channel exists
func (r *ChannelRepositoryImpl) Exists(ctx context.Context, id *channel.ChannelID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChannelModel{}).
		Where("id = ? AND deleted_at IS NULL", id.String()).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check channel existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a channel with the given name exists
func (r *ChannelRepositoryImpl) ExistsByName(ctx context.Context, name *channel.ChannelName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChannelModel{}).
		Where("name = ? AND deleted_at IS NULL", name.String()).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check channel name existence: %w", err)
	}

	return count > 0, nil
}

// toChannelModel converts domain channel to GORM model
func (r *ChannelRepositoryImpl) toChannelModel(ch *channel.Channel) (*models.ChannelModel, error) {
	// Convert config to JSON
	config := models.JSON(ch.Config().ToMap())

	// Convert recipients to JSONArray
	var recipients models.JSONArray
	recipientSlice := ch.Recipients().ToSlice()
	recipientData, err := json.Marshal(recipientSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipients: %w", err)
	}
	if err := json.Unmarshal(recipientData, &recipients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients to JSONArray type: %w", err)
	}

	// Handle template ID
	var templateID *string
	if ch.TemplateID() != nil {
		id := ch.TemplateID().String()
		templateID = &id
	}

	// Handle deleted_at
	var deletedAt *int64
	if ch.Timestamps().DeletedAt != nil {
		deletedAt = ch.Timestamps().DeletedAt
	}

	return &models.ChannelModel{
		ID:            ch.ID().String(),
		Name:          ch.Name().String(),
		Description:   ch.Description().String(),
		Enabled:       ch.IsEnabled(),
		ChannelType:   ch.ChannelType().String(),
		TemplateID:    templateID,
		Timeout:       ch.CommonSettings().Timeout,
		RetryAttempts: ch.CommonSettings().RetryAttempts,
		RetryDelay:    ch.CommonSettings().RetryDelay,
		Config:        config,
		Recipients:    recipients,
		Tags:          pq.StringArray(ch.Tags().ToSlice()),
		CreatedAt:     ch.Timestamps().CreatedAt,
		UpdatedAt:     ch.Timestamps().UpdatedAt,
		DeletedAt:     deletedAt,
		LastUsed:      ch.LastUsed(),
	}, nil
}

// fromChannelModel converts GORM model to domain channel
func (r *ChannelRepositoryImpl) fromChannelModel(model *models.ChannelModel) (*channel.Channel, error) {
	// Convert ID
	id, err := channel.NewChannelIDFromString(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// Convert name
	name, err := channel.NewChannelName(model.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}

	// Convert description
	description, err := channel.NewDescription(model.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Convert channel type
	channelType, err := shared.NewChannelTypeFromString(model.ChannelType)
	if err != nil {
		return nil, fmt.Errorf("invalid channel type: %s, error: %w", model.ChannelType, err)
	}

	// Convert template ID
	var templateID *template.TemplateID
	if model.TemplateID != nil {
		templateID, err = template.NewTemplateIDFromString(*model.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
	}

	// Convert common settings
	commonSettings, err := shared.NewCommonSettings(model.Timeout, model.RetryAttempts, model.RetryDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid common settings: %w", err)
	}

	// Convert config
	configMap := map[string]interface{}(model.Config)
	config := channel.NewChannelConfig(configMap)

	// Convert recipients
	var recipientSlice []*channel.Recipient
	recipientData, err := json.Marshal(model.Recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipients: %w", err)
	}
	if err := json.Unmarshal(recipientData, &recipientSlice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
	}
	recipients := channel.NewRecipients(recipientSlice)

	// Convert tags
	tags := channel.NewTags(model.Tags)

	// Convert timestamps
	timestamps := &shared.Timestamps{
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		DeletedAt: model.DeletedAt,
	}

	// Reconstruct channel
	return channel.ReconstructChannel(
		id,
		name,
		description,
		model.Enabled,
		channelType,
		templateID,
		commonSettings,
		config,
		recipients,
		tags,
		timestamps,
		model.LastUsed,
	), nil
}
