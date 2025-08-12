package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"notification/internal/domain/channel"
	"notification/internal/domain/message"
)

// MessageRepositoryImpl implements message.MessageRepository interface
type MessageRepositoryImpl struct {
	db *sqlx.DB
}

// NewMessageRepositoryImpl creates a new message repository implementation
func NewMessageRepositoryImpl(db *sqlx.DB) *MessageRepositoryImpl {
	return &MessageRepositoryImpl{
		db: db,
	}
}

// messageRow represents message data in database
type messageRow struct {
	ID               string `db:"id"`
	ChannelIDs       string `db:"channel_ids"`       // JSON array
	Variables        string `db:"variables"`         // JSON object
	ChannelOverrides string `db:"channel_overrides"` // JSON object
	Status           string `db:"status"`
	CreatedAt        int64  `db:"created_at"`
}

// messageResultRow represents message result data in database
type messageResultRow struct {
	ID           int            `db:"id"`
	MessageID    string         `db:"message_id"`
	ChannelID    string         `db:"channel_id"`
	Status       string         `db:"status"`
	Message      string         `db:"message"`
	ErrorCode    sql.NullString `db:"error_code"`
	ErrorDetails sql.NullString `db:"error_details"`
	SentAt       sql.NullInt64  `db:"sent_at"`
}

// Save saves a message to the database
func (r *MessageRepositoryImpl) Save(ctx context.Context, msg *message.Message) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Convert message to row
	row, err := r.toMessageRow(msg)
	if err != nil {
		return fmt.Errorf("failed to convert message to row: %w", err)
	}

	// Insert message
	messageQuery := `
		INSERT INTO messages (
			id, channel_ids, variables, channel_overrides, status, created_at
		) VALUES (
			:id, :channel_ids, :variables, :channel_overrides, :status, :created_at
		)`

	_, err = tx.NamedExecContext(ctx, messageQuery, row)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Insert message results
	for _, result := range msg.Results() {
		resultRow, err := r.toMessageResultRow(msg.ID(), result)
		if err != nil {
			return fmt.Errorf("failed to convert message result to row: %w", err)
		}

		resultQuery := `
			INSERT INTO message_results (
				message_id, channel_id, status, message, error_code, error_details, sent_at
			) VALUES (
				:message_id, :channel_id, :status, :message, :error_code, :error_details, :sent_at
			)`

		_, err = tx.NamedExecContext(ctx, resultQuery, resultRow)
		if err != nil {
			return fmt.Errorf("failed to save message result: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID finds a message by its ID
func (r *MessageRepositoryImpl) FindByID(ctx context.Context, id *message.MessageID) (*message.Message, error) {
	// Find message
	messageQuery := `
		SELECT id, channel_ids, variables, channel_overrides, status, created_at
		FROM messages 
		WHERE id = $1`

	var row messageRow
	err := r.db.GetContext(ctx, &row, messageQuery, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	// Find message results
	resultsQuery := `
		SELECT id, message_id, channel_id, status, message, error_code, error_details, sent_at
		FROM message_results 
		WHERE message_id = $1
		ORDER BY id`

	var resultRows []messageResultRow
	err = r.db.SelectContext(ctx, &resultRows, resultsQuery, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find message results: %w", err)
	}

	return r.fromMessageRow(&row, resultRows)
}

// Update updates a message in the database
func (r *MessageRepositoryImpl) Update(ctx context.Context, msg *message.Message) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Convert message to row
	row, err := r.toMessageRow(msg)
	if err != nil {
		return fmt.Errorf("failed to convert message to row: %w", err)
	}

	// Update message
	messageQuery := `
		UPDATE messages SET
			channel_ids = :channel_ids,
			variables = :variables,
			channel_overrides = :channel_overrides,
			status = :status
		WHERE id = :id`

	_, err = tx.NamedExecContext(ctx, messageQuery, row)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	// Delete existing results
	_, err = tx.ExecContext(ctx, "DELETE FROM message_results WHERE message_id = $1", msg.ID().String())
	if err != nil {
		return fmt.Errorf("failed to delete existing message results: %w", err)
	}

	// Insert updated results
	for _, result := range msg.Results() {
		resultRow, err := r.toMessageResultRow(msg.ID(), result)
		if err != nil {
			return fmt.Errorf("failed to convert message result to row: %w", err)
		}

		resultQuery := `
			INSERT INTO message_results (
				message_id, channel_id, status, message, error_code, error_details, sent_at
			) VALUES (
				:message_id, :channel_id, :status, :message, :error_code, :error_details, :sent_at
			)`

		_, err = tx.NamedExecContext(ctx, resultQuery, resultRow)
		if err != nil {
			return fmt.Errorf("failed to save message result: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Exists checks if a message exists
func (r *MessageRepositoryImpl) Exists(ctx context.Context, id *message.MessageID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id.String())
	if err != nil {
		return false, fmt.Errorf("failed to check message existence: %w", err)
	}

	return exists, nil
}

// toMessageRow converts domain message to database row
func (r *MessageRepositoryImpl) toMessageRow(msg *message.Message) (*messageRow, error) {
	// Convert channel IDs to JSON
	channelIDStrings := make([]string, 0, msg.ChannelIDs().Count())
	for _, id := range msg.ChannelIDs().ToSlice() {
		channelIDStrings = append(channelIDStrings, id.String())
	}
	channelIDsJSON, err := json.Marshal(channelIDStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channel IDs: %w", err)
	}

	// Convert variables to JSON
	variablesJSON, err := json.Marshal(msg.Variables().ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	// Convert channel overrides to JSON
	channelOverridesJSON, err := json.Marshal(msg.ChannelOverrides().ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channel overrides: %w", err)
	}

	return &messageRow{
		ID:               msg.ID().String(),
		ChannelIDs:       string(channelIDsJSON),
		Variables:        string(variablesJSON),
		ChannelOverrides: string(channelOverridesJSON),
		Status:           string(msg.Status()),
		CreatedAt:        msg.CreatedAt(),
	}, nil
}

// toMessageResultRow converts domain message result to database row
func (r *MessageRepositoryImpl) toMessageResultRow(messageID *message.MessageID, result *message.MessageResult) (*messageResultRow, error) {
	row := &messageResultRow{
		MessageID: messageID.String(),
		ChannelID: result.ChannelID().String(),
		Status:    string(result.Status()),
		Message:   result.Message(),
	}

	// Handle error
	if result.Error() != nil {
		row.ErrorCode = sql.NullString{
			String: result.Error().Code,
			Valid:  true,
		}
		row.ErrorDetails = sql.NullString{
			String: result.Error().Details,
			Valid:  true,
		}
	}

	// Handle sent_at
	if result.SentAt() != nil {
		row.SentAt = sql.NullInt64{
			Int64: *result.SentAt(),
			Valid: true,
		}
	}

	return row, nil
}

// fromMessageRow converts database row to domain message
func (r *MessageRepositoryImpl) fromMessageRow(row *messageRow, resultRows []messageResultRow) (*message.Message, error) {
	// Convert ID
	id, err := message.NewMessageIDFromString(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Convert channel IDs
	var channelIDStrings []string
	if err := json.Unmarshal([]byte(row.ChannelIDs), &channelIDStrings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel IDs: %w", err)
	}

	channelIDs := make([]*channel.ChannelID, 0, len(channelIDStrings))
	for _, idStr := range channelIDStrings {
		channelID, err := channel.NewChannelIDFromString(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid channel ID: %w", err)
		}
		channelIDs = append(channelIDs, channelID)
	}

	channelIDsVO, err := message.NewChannelIDs(channelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel IDs: %w", err)
	}

	// Convert variables
	var variablesMap map[string]interface{}
	if err := json.Unmarshal([]byte(row.Variables), &variablesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}
	variables := message.NewVariables(variablesMap)

	// Convert channel overrides
	var channelOverridesMap map[string]*message.ChannelOverride
	if err := json.Unmarshal([]byte(row.ChannelOverrides), &channelOverridesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel overrides: %w", err)
	}
	channelOverrides := message.NewChannelOverrides(channelOverridesMap)

	// Convert status
	status := message.MessageStatus(row.Status)

	// Convert results
	results := make([]*message.MessageResult, 0, len(resultRows))
	for _, resultRow := range resultRows {
		result, err := r.fromMessageResultRow(&resultRow)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message result: %w", err)
		}
		results = append(results, result)
	}

	// Reconstruct message
	return message.ReconstructMessage(
		id,
		channelIDsVO,
		variables,
		channelOverrides,
		status,
		results,
		row.CreatedAt,
	), nil
}

// fromMessageResultRow converts database row to domain message result
func (r *MessageRepositoryImpl) fromMessageResultRow(row *messageResultRow) (*message.MessageResult, error) {
	// Convert channel ID
	channelID, err := channel.NewChannelIDFromString(row.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// Convert status and create result
	status := message.MessageResultStatus(row.Status)
	if status == message.MessageResultStatusSuccess {
		return message.NewSuccessfulMessageResult(channelID, row.Message)
	} else {
		// Handle error
		var msgError *message.MessageError
		if row.ErrorCode.Valid {
			msgError = message.NewMessageError(row.ErrorCode.String, row.ErrorDetails.String)
		} else {
			msgError = message.NewMessageError("UNKNOWN_ERROR", "Unknown error occurred")
		}

		return message.NewFailedMessageResult(channelID, row.Message, msgError)
	}
}