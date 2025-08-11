package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"channel-api/internal/domain/channel"
	"channel-api/internal/domain/shared"
	"channel-api/internal/domain/template"
)

// ChannelRepositoryImpl implements channel.ChannelRepository interface
type ChannelRepositoryImpl struct {
	db *sqlx.DB
}

// NewChannelRepositoryImpl creates a new channel repository implementation
func NewChannelRepositoryImpl(db *sqlx.DB) *ChannelRepositoryImpl {
	return &ChannelRepositoryImpl{
		db: db,
	}
}

// channelRow represents channel data in database
type channelRow struct {
	ID             string         `db:"id"`
	Name           string         `db:"name"`
	Description    string         `db:"description"`
	Enabled        bool           `db:"enabled"`
	ChannelType    string         `db:"channel_type"`
	TemplateID     sql.NullString `db:"template_id"`
	Timeout        int            `db:"timeout"`
	RetryAttempts  int            `db:"retry_attempts"`
	RetryDelay     int            `db:"retry_delay"`
	Config         string         `db:"config"`         // JSON string
	Recipients     string         `db:"recipients"`     // JSON string
	Tags           pq.StringArray `db:"tags"`
	CreatedAt      int64          `db:"created_at"`
	UpdatedAt      int64          `db:"updated_at"`
	DeletedAt      sql.NullInt64  `db:"deleted_at"`
	LastUsed       sql.NullInt64  `db:"last_used"`
}

// Save saves a channel to the database
func (r *ChannelRepositoryImpl) Save(ctx context.Context, ch *channel.Channel) error {
	row, err := r.toChannelRow(ch)
	if err != nil {
		return fmt.Errorf("failed to convert channel to row: %w", err)
	}

	query := `
		INSERT INTO channels (
			id, name, description, enabled, channel_type, template_id,
			timeout, retry_attempts, retry_delay, config, recipients, tags,
			created_at, updated_at, deleted_at, last_used
		) VALUES (
			:id, :name, :description, :enabled, :channel_type, :template_id,
			:timeout, :retry_attempts, :retry_delay, :config, :recipients, :tags,
			:created_at, :updated_at, :deleted_at, :last_used
		)`

	_, err = r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to save channel: %w", err)
	}

	return nil
}

// FindByID finds a channel by its ID
func (r *ChannelRepositoryImpl) FindByID(ctx context.Context, id *channel.ChannelID) (*channel.Channel, error) {
	query := `
		SELECT id, name, description, enabled, channel_type, template_id,
			   timeout, retry_attempts, retry_delay, config, recipients, tags,
			   created_at, updated_at, deleted_at, last_used
		FROM channels 
		WHERE id = $1 AND deleted_at IS NULL`

	var row channelRow
	err := r.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	return r.fromChannelRow(&row)
}

// FindByName finds a channel by its name
func (r *ChannelRepositoryImpl) FindByName(ctx context.Context, name *channel.ChannelName) (*channel.Channel, error) {
	query := `
		SELECT id, name, description, enabled, channel_type, template_id,
			   timeout, retry_attempts, retry_delay, config, recipients, tags,
			   created_at, updated_at, deleted_at, last_used
		FROM channels 
		WHERE name = $1 AND deleted_at IS NULL`

	var row channelRow
	err := r.db.GetContext(ctx, &row, query, name.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	return r.fromChannelRow(&row)
}

// FindAll finds all channels with filtering and pagination
func (r *ChannelRepositoryImpl) FindAll(ctx context.Context, filter *channel.ChannelFilter, pagination *shared.Pagination) (*shared.PaginatedResult[*channel.Channel], error) {
	// Build WHERE clause
	whereClause, args, err := r.buildWhereClause(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %w", err)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM channels %s", whereClause)
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to count channels: %w", err)
	}

	// Query channels with pagination
	query := fmt.Sprintf(`
		SELECT id, name, description, enabled, channel_type, template_id,
			   timeout, retry_attempts, retry_delay, config, recipients, tags,
			   created_at, updated_at, deleted_at, last_used
		FROM channels %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pagination.MaxResultCount, pagination.SkipCount)

	var rows []channelRow
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}

	// Convert to domain objects
	channels := make([]*channel.Channel, 0, len(rows))
	for _, row := range rows {
		ch, err := r.fromChannelRow(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to channel: %w", err)
		}
		channels = append(channels, ch)
	}

	// Calculate hasMore
	hasMore := pagination.SkipCount+len(channels) < totalCount

	return &shared.PaginatedResult[*channel.Channel]{
		Items:          channels,
		SkipCount:      pagination.SkipCount,
		MaxResultCount: pagination.MaxResultCount,
		TotalCount:     totalCount,
		HasMore:        hasMore,
	}, nil
}

// Update updates a channel in the database
func (r *ChannelRepositoryImpl) Update(ctx context.Context, ch *channel.Channel) error {
	row, err := r.toChannelRow(ch)
	if err != nil {
		return fmt.Errorf("failed to convert channel to row: %w", err)
	}

	query := `
		UPDATE channels SET
			name = :name,
			description = :description,
			enabled = :enabled,
			channel_type = :channel_type,
			template_id = :template_id,
			timeout = :timeout,
			retry_attempts = :retry_attempts,
			retry_delay = :retry_delay,
			config = :config,
			recipients = :recipients,
			tags = :tags,
			updated_at = :updated_at,
			deleted_at = :deleted_at,
			last_used = :last_used
		WHERE id = :id`

	_, err = r.db.NamedExecContext(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	return nil
}

// Delete deletes a channel from the database (hard delete)
func (r *ChannelRepositoryImpl) Delete(ctx context.Context, id *channel.ChannelID) error {
	query := `DELETE FROM channels WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// Exists checks if a channel exists
func (r *ChannelRepositoryImpl) Exists(ctx context.Context, id *channel.ChannelID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM channels WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id.String())
	if err != nil {
		return false, fmt.Errorf("failed to check channel existence: %w", err)
	}

	return exists, nil
}

// ExistsByName checks if a channel with the given name exists
func (r *ChannelRepositoryImpl) ExistsByName(ctx context.Context, name *channel.ChannelName) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM channels WHERE name = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, name.String())
	if err != nil {
		return false, fmt.Errorf("failed to check channel name existence: %w", err)
	}

	return exists, nil
}

// buildWhereClause builds WHERE clause for filtering
func (r *ChannelRepositoryImpl) buildWhereClause(filter *channel.ChannelFilter) (string, []interface{}, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if filter.HasChannelTypeFilter() {
		conditions = append(conditions, fmt.Sprintf("channel_type = $%d", argIndex))
		args = append(args, string(*filter.ChannelType))
		argIndex++
	}

	if filter.HasTagsFilter() {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argIndex))
		args = append(args, pq.Array(filter.Tags))
		argIndex++
	}

	if filter.HasEnabledFilter() {
		conditions = append(conditions, fmt.Sprintf("enabled = $%d", argIndex))
		args = append(args, *filter.Enabled)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args, nil
}

// toChannelRow converts domain channel to database row
func (r *ChannelRepositoryImpl) toChannelRow(ch *channel.Channel) (*channelRow, error) {
	// Convert config to JSON
	configJSON, err := json.Marshal(ch.Config().ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Convert recipients to JSON
	recipientsJSON, err := json.Marshal(ch.Recipients().ToSlice())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipients: %w", err)
	}

	// Handle template ID
	var templateID sql.NullString
	if ch.TemplateID() != nil {
		templateID = sql.NullString{
			String: ch.TemplateID().String(),
			Valid:  true,
		}
	}

	// Handle deleted_at
	var deletedAt sql.NullInt64
	if ch.Timestamps().DeletedAt != nil {
		deletedAt = sql.NullInt64{
			Int64: *ch.Timestamps().DeletedAt,
			Valid: true,
		}
	}

	// Handle last_used
	var lastUsed sql.NullInt64
	if ch.LastUsed() != nil {
		lastUsed = sql.NullInt64{
			Int64: *ch.LastUsed(),
			Valid: true,
		}
	}

	return &channelRow{
		ID:            ch.ID().String(),
		Name:          ch.Name().String(),
		Description:   ch.Description().String(),
		Enabled:       ch.IsEnabled(),
		ChannelType:   string(ch.ChannelType()),
		TemplateID:    templateID,
		Timeout:       ch.CommonSettings().Timeout,
		RetryAttempts: ch.CommonSettings().RetryAttempts,
		RetryDelay:    ch.CommonSettings().RetryDelay,
		Config:        string(configJSON),
		Recipients:    string(recipientsJSON),
		Tags:          pq.StringArray(ch.Tags().ToSlice()),
		CreatedAt:     ch.Timestamps().CreatedAt,
		UpdatedAt:     ch.Timestamps().UpdatedAt,
		DeletedAt:     deletedAt,
		LastUsed:      lastUsed,
	}, nil
}

// fromChannelRow converts database row to domain channel
func (r *ChannelRepositoryImpl) fromChannelRow(row *channelRow) (*channel.Channel, error) {
	// Convert ID
	id, err := channel.NewChannelIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// Convert name
	name, err := channel.NewChannelName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}

	// Convert description
	description, err := channel.NewDescription(row.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Convert channel type
	channelType := shared.ChannelType(row.ChannelType)
	if !channelType.IsValid() {
		return nil, fmt.Errorf("invalid channel type: %s", row.ChannelType)
	}

	// Convert template ID
	var templateID *template.TemplateID
	if row.TemplateID.Valid {
		templateID, err = template.NewTemplateIDFromString(row.TemplateID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid template ID: %w", err)
		}
	}

	// Convert common settings
	commonSettings, err := shared.NewCommonSettings(row.Timeout, row.RetryAttempts, row.RetryDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid common settings: %w", err)
	}

	// Convert config
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(row.Config), &configMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	config := channel.NewChannelConfig(configMap)

	// Convert recipients
	var recipientSlice []*channel.Recipient
	if err := json.Unmarshal([]byte(row.Recipients), &recipientSlice); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recipients: %w", err)
	}
	recipients := channel.NewRecipients(recipientSlice)

	// Convert tags
	tags := channel.NewTags([]string(row.Tags))

	// Convert timestamps
	timestamps := &shared.Timestamps{
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		timestamps.DeletedAt = &row.DeletedAt.Int64
	}

	// Convert last used
	var lastUsed *int64
	if row.LastUsed.Valid {
		lastUsed = &row.LastUsed.Int64
	}

	// Reconstruct channel
	return channel.ReconstructChannel(
		id,
		name,
		description,
		row.Enabled,
		channelType,
		templateID,
		commonSettings,
		config,
		recipients,
		tags,
		timestamps,
		lastUsed,
	), nil
}