-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) DEFAULT '',
    channel_type VARCHAR(50) NOT NULL,
    subject VARCHAR(200) DEFAULT '',
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_templates_name ON templates(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_templates_type ON templates(channel_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_templates_tags ON templates USING GIN(tags) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_templates_created_at ON templates(created_at) WHERE deleted_at IS NULL;

-- Add constraint for channel_type
ALTER TABLE templates ADD CONSTRAINT check_template_channel_type 
    CHECK (channel_type IN ('email', 'slack', 'sms'));

-- Add constraint for version
ALTER TABLE templates ADD CONSTRAINT check_template_version 
    CHECK (version > 0);

-- Add unique constraint on name for non-deleted templates
CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_name_unique 
    ON templates(name) WHERE deleted_at IS NULL;