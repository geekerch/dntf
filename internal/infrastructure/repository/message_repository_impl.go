package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"notification/internal/domain/channel"
	"notification/internal/domain/message"
	"notification/internal/infrastructure/models"
)

// MessageRepositoryImpl implements message.MessageRepository interface using GORM
type MessageRepositoryImpl struct {
	db *gorm.DB
}

// NewMessageRepositoryImpl creates a new message repository implementation
func NewMessageRepositoryImpl(db *gorm.DB) *MessageRepositoryImpl {
	return &MessageRepositoryImpl{
		db: db,
	}
}

// Save saves a message to the database
func (r *MessageRepositoryImpl) Save(ctx context.Context, msg *message.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Convert message to model
		messageModel, err := r.toMessageModel(msg)
		if err != nil {
			return fmt.Errorf("failed to convert message to model: %w", err)
		}

		// Save message
		if err := tx.Create(messageModel).Error; err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		// Save message results
		for _, result := range msg.Results() {
			resultModel, err := r.toMessageResultModel(msg.ID(), result)
			if err != nil {
				return fmt.Errorf("failed to convert message result to model: %w", err)
			}

			if err := tx.Create(resultModel).Error; err != nil {
				return fmt.Errorf("failed to save message result: %w", err)
			}
		}

		return nil
	})
}

// FindByID finds a message by its ID
func (r *MessageRepositoryImpl) FindByID(ctx context.Context, id *message.MessageID) (*message.Message, error) {
	var messageModel models.MessageModel
	
	err := r.db.WithContext(ctx).
		Preload("Results").
		Where("id = ?", id.String()).
		First(&messageModel).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	return r.fromMessageModel(&messageModel)
}

// Update updates a message in the database
func (r *MessageRepositoryImpl) Update(ctx context.Context, msg *message.Message) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Convert message to model
		messageModel, err := r.toMessageModel(msg)
		if err != nil {
			return fmt.Errorf("failed to convert message to model: %w", err)
		}

		// Update message
		if err := tx.Save(messageModel).Error; err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}

		// Delete existing results
		if err := tx.Where("message_id = ?", msg.ID().String()).Delete(&models.MessageResultModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing message results: %w", err)
		}

		// Save updated results
		for _, result := range msg.Results() {
			resultModel, err := r.toMessageResultModel(msg.ID(), result)
			if err != nil {
				return fmt.Errorf("failed to convert message result to model: %w", err)
			}

			if err := tx.Create(resultModel).Error; err != nil {
				return fmt.Errorf("failed to save message result: %w", err)
			}
		}

		return nil
	})
}

// Exists checks if a message exists
func (r *MessageRepositoryImpl) Exists(ctx context.Context, id *message.MessageID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.MessageModel{}).
		Where("id = ?", id.String()).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("failed to check message existence: %w", err)
	}

	return count > 0, nil
}

// toMessageModel converts domain message to GORM model
func (r *MessageRepositoryImpl) toMessageModel(msg *message.Message) (*models.MessageModel, error) {
	// Convert channel IDs to JSONArray
	channelIDStrings := make([]string, 0, msg.ChannelIDs().Count())
	for _, id := range msg.ChannelIDs().ToSlice() {
		channelIDStrings = append(channelIDStrings, id.String())
	}
	// Convert string array to JSONArray (array of maps)
	channelIDs := make(models.JSONArray, len(channelIDStrings))
	for i, idStr := range channelIDStrings {
		channelIDs[i] = map[string]interface{}{"id": idStr}
	}

	// Convert variables to JSON
	variables := models.JSON(msg.Variables().ToMap())

	// Convert channel overrides to JSON
	channelOverrides := models.JSON{}
	overrideData, err := json.Marshal(msg.ChannelOverrides().ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channel overrides: %w", err)
	}
	if err := json.Unmarshal(overrideData, &channelOverrides); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel overrides to JSON type: %w", err)
	}

	return &models.MessageModel{
		ID:               msg.ID().String(),
		ChannelIDs:       channelIDs,
		Variables:        variables,
		ChannelOverrides: channelOverrides,
		Status:           string(msg.Status()),
		CreatedAt:        msg.CreatedAt(),
	}, nil
}

// toMessageResultModel converts domain message result to GORM model
func (r *MessageRepositoryImpl) toMessageResultModel(messageID *message.MessageID, result *message.MessageResult) (*models.MessageResultModel, error) {
	model := &models.MessageResultModel{
		MessageID: messageID.String(),
		ChannelID: result.ChannelID().String(),
		Status:    string(result.Status()),
		Message:   result.Message(),
		SentAt:    result.SentAt(),
	}

	// Handle error
	if result.Error() != nil {
		errorCode := result.Error().Code
		errorDetails := result.Error().Details
		model.ErrorCode = &errorCode
		model.ErrorDetails = &errorDetails
	}

	return model, nil
}

// fromMessageModel converts GORM model to domain message
func (r *MessageRepositoryImpl) fromMessageModel(model *models.MessageModel) (*message.Message, error) {
	// Convert ID
	id, err := message.NewMessageIDFromString(model.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Convert channel IDs from JSONArray
	channelIDStrings := make([]string, len(model.ChannelIDs))
	for i, idMap := range model.ChannelIDs {
		if id, ok := idMap["id"].(string); ok {
			channelIDStrings[i] = id
		} else {
			return nil, fmt.Errorf("invalid channel ID format in database")
		}
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
	variablesMap := map[string]interface{}(model.Variables)
	variables := message.NewVariables(variablesMap)

	// Convert channel overrides
	var channelOverridesMap map[string]*message.ChannelOverride
	overrideData, err := json.Marshal(model.ChannelOverrides)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channel overrides: %w", err)
	}
	if err := json.Unmarshal(overrideData, &channelOverridesMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel overrides: %w", err)
	}
	channelOverrides := message.NewChannelOverrides(channelOverridesMap)

	// Convert status
	status := message.MessageStatus(model.Status)

	// Convert results
	results := make([]*message.MessageResult, 0, len(model.Results))
	for _, resultModel := range model.Results {
		result, err := r.fromMessageResultModel(&resultModel)
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
		model.CreatedAt,
	), nil
}

// fromMessageResultModel converts GORM model to domain message result
func (r *MessageRepositoryImpl) fromMessageResultModel(model *models.MessageResultModel) (*message.MessageResult, error) {
	// Convert channel ID
	channelID, err := channel.NewChannelIDFromString(model.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// Convert status and create result
	status := message.MessageResultStatus(model.Status)
	if status == message.MessageResultStatusSuccess {
		return message.NewSuccessfulMessageResult(channelID, model.Message)
	} else {
		// Handle error
		var msgError *message.MessageError
		if model.ErrorCode != nil {
			errorDetails := ""
			if model.ErrorDetails != nil {
				errorDetails = *model.ErrorDetails
			}
			msgError = message.NewMessageError(*model.ErrorCode, errorDetails)
		} else {
			msgError = message.NewMessageError("UNKNOWN_ERROR", "Unknown error occurred")
		}

		return message.NewFailedMessageResult(channelID, model.Message, msgError)
	}
}