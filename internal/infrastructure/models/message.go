package models

import (
	"gorm.io/gorm"
)

// MessageModel represents the message table structure for GORM
type MessageModel struct {
	ID               string             `gorm:"primaryKey;type:varchar(255)" json:"id"`
	ChannelIDs       JSONArray          `gorm:"type:jsonb;not null" json:"channel_ids"`
	Variables        JSON               `gorm:"type:jsonb;not null" json:"variables"`
	ChannelOverrides JSON               `gorm:"type:jsonb;not null;default:'{}'" json:"channel_overrides"`
	Status           string             `gorm:"type:varchar(50);not null;default:'pending';index:idx_messages_status;check:status IN ('pending','success','failed','partial_success')" json:"status"`
	CreatedAt        int64              `gorm:"not null;index:idx_messages_created_at" json:"created_at"`
	Results          []MessageResultModel `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"results,omitempty"`
}

// TableName returns the table name for GORM
func (MessageModel) TableName() string {
	return "messages"
}

// BeforeCreate GORM hook
func (m *MessageModel) BeforeCreate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// BeforeUpdate GORM hook
func (m *MessageModel) BeforeUpdate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// MessageResultModel represents the message_results table structure for GORM
type MessageResultModel struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID    string `gorm:"type:varchar(255);not null;index:idx_message_results_message_id;uniqueIndex:idx_message_results_unique,priority:1" json:"message_id"`
	ChannelID    string `gorm:"type:varchar(255);not null;index:idx_message_results_channel_id;uniqueIndex:idx_message_results_unique,priority:2" json:"channel_id"`
	Status       string `gorm:"type:varchar(50);not null;index:idx_message_results_status;check:status IN ('success','failed')" json:"status"`
	Message      string `gorm:"type:text;not null" json:"message"`
	ErrorCode    *string `gorm:"type:varchar(100)" json:"error_code"`
	ErrorDetails *string `gorm:"type:text" json:"error_details"`
	SentAt       *int64  `json:"sent_at"`
	
	// Foreign key relationship
	MessageModel MessageModel `gorm:"foreignKey:MessageID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName returns the table name for GORM
func (MessageResultModel) TableName() string {
	return "message_results"
}

// BeforeCreate GORM hook
func (mr *MessageResultModel) BeforeCreate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// BeforeUpdate GORM hook
func (mr *MessageResultModel) BeforeUpdate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}