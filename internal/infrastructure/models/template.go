package models

import (
	"gorm.io/gorm"
)

// TemplateModel represents the template table structure for GORM
type TemplateModel struct {
	ID          string      `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name        string      `gorm:"type:varchar(100);not null;uniqueIndex:idx_templates_name_unique,where:deleted_at IS NULL" json:"name"`
	Description string      `gorm:"type:varchar(500);default:''" json:"description"`
	ChannelType string      `gorm:"type:varchar(50);not null;index:idx_templates_type,where:deleted_at IS NULL;check:channel_type IN ('email','slack','sms')" json:"channel_type"`
	Subject     string      `gorm:"type:varchar(200);default:''" json:"subject"`
	Content     string      `gorm:"type:text;not null" json:"content"`
	Tags        StringArray `gorm:"type:text[];default:'{}';index:idx_templates_tags,type:gin,where:deleted_at IS NULL" json:"tags"`
	CreatedAt   int64       `gorm:"not null;index:idx_templates_created_at,where:deleted_at IS NULL" json:"created_at"`
	UpdatedAt   int64       `gorm:"not null" json:"updated_at"`
	DeletedAt   *int64      `gorm:"index" json:"deleted_at"`
	Version     int         `gorm:"not null;default:1;check:version > 0" json:"version"`
}

// TableName returns the table name for GORM
func (TemplateModel) TableName() string {
	return "templates"
}

// BeforeCreate GORM hook
func (t *TemplateModel) BeforeCreate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}

// BeforeUpdate GORM hook
func (t *TemplateModel) BeforeUpdate(tx *gorm.DB) error {
	// Additional validation can be added here
	return nil
}