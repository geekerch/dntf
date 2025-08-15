package models

import (
	"gorm.io/gorm"
)

// AllModels returns all GORM models for migration
func AllModels() []interface{} {
	return []interface{}{
		&ChannelModel{},
		&TemplateModel{},
		&MessageModel{},
		&MessageResultModel{},
	}
}

// MigrateModels runs GORM AutoMigrate for all models
func MigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}

// CreateIndexes creates additional indexes that GORM might not handle automatically
func CreateIndexes(db *gorm.DB) error {
	// PostgreSQL specific indexes
	if db.Dialector.Name() == "postgres" {
		// Create GIN indexes for array fields (PostgreSQL specific)
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_channels_tags_gin ON channels USING GIN(tags) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
		
		if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_templates_tags_gin ON templates USING GIN(tags) WHERE deleted_at IS NULL").Error; err != nil {
			return err
		}
	}
	
	return nil
}

// DropTables drops all tables (useful for testing)
func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(AllModels()...)
}