package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ChannelModel represents the channel table structure for GORM
type ChannelModel struct {
	ID            string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name          string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_channels_name_unique,where:deleted_at IS NULL" json:"name"`
	Description   string         `gorm:"type:varchar(500);default:''" json:"description"`
	Enabled       bool           `gorm:"not null;default:true;index:idx_channels_enabled,where:deleted_at IS NULL" json:"enabled"`
	ChannelType   string         `gorm:"type:varchar(50);not null;index:idx_channels_type,where:deleted_at IS NULL;check:channel_type IN ('email','slack','sms')" json:"channel_type"`
	TemplateID    *string        `gorm:"type:varchar(255);index:idx_channels_template_id,where:deleted_at IS NULL" json:"template_id"`
	Timeout       int            `gorm:"not null;check:timeout > 0" json:"timeout"`
	RetryAttempts int            `gorm:"not null;default:0;check:retry_attempts >= 0" json:"retry_attempts"`
	RetryDelay    int            `gorm:"not null;default:0;check:retry_delay >= 0" json:"retry_delay"`
	Config        JSON           `gorm:"type:jsonb;not null" json:"config"`
	Recipients    JSONArray      `gorm:"type:jsonb;not null" json:"recipients"`
	Tags          pq.StringArray `gorm:"type:text[];default:'{}'" json:"tags"`
	CreatedAt     int64          `gorm:"not null;index:idx_channels_created_at,where:deleted_at IS NULL" json:"created_at"`
	UpdatedAt     int64          `gorm:"not null" json:"updated_at"`
	DeletedAt     *int64         `gorm:"index" json:"deleted_at"`
	LastUsed      *int64         `json:"last_used"`
}

// TableName returns the table name for GORM
func (ChannelModel) TableName() string {
	return "channels"
}

// BeforeCreate GORM hook
func (c *ChannelModel) BeforeCreate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// BeforeUpdate GORM hook
func (c *ChannelModel) BeforeUpdate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// JSON is a custom type for handling JSON objects
type JSON map[string]interface{}

// Scan implements the Scanner interface for database/sql
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// JSONArray is a custom type for handling JSON arrays of objects
type JSONArray []map[string]interface{}

// Scan implements the Scanner interface for database/sql
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = []map[string]interface{}{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	// Try to unmarshal as an array of maps
	if err := json.Unmarshal(bytes, j); err == nil {
		return nil
	}

	// If it's not an array, try to unmarshal as a single map and convert to an array
	var singleMap map[string]interface{}
	if err := json.Unmarshal(bytes, &singleMap); err == nil {
		*j = []map[string]interface{}{singleMap}
		return nil
	}

	return fmt.Errorf("failed to unmarshal JSONArray: %w", errors.New(string(bytes)))
}

// Value implements the driver Valuer interface
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
