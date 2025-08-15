-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    channel_type VARCHAR(50) NOT NULL,
    template_id VARCHAR(255),
    timeout INTEGER NOT NULL,
    retry_attempts INTEGER NOT NULL DEFAULT 0,
    retry_delay INTEGER NOT NULL DEFAULT 0,
    config JSONB NOT NULL,
    recipients JSONB NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT,
    last_used BIGINT
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_channels_name ON channels(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(channel_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_enabled ON channels(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_tags ON channels USING GIN(tags) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_created_at ON channels(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_template_id ON channels(template_id) WHERE deleted_at IS NULL;

-- Add constraint for channel_type
ALTER TABLE channels ADD CONSTRAINT check_channel_type 
    CHECK (channel_type IN ('email', 'slack', 'sms'));

-- Add constraint for timeout
ALTER TABLE channels ADD CONSTRAINT check_timeout 
    CHECK (timeout > 0);

-- Add constraint for retry_attempts
ALTER TABLE channels ADD CONSTRAINT check_retry_attempts 
    CHECK (retry_attempts >= 0);

-- Add constraint for retry_delay
ALTER TABLE channels ADD CONSTRAINT check_retry_delay 
    CHECK (retry_delay >= 0);

-- Add unique constraint on name for non-deleted channels
CREATE UNIQUE INDEX IF NOT EXISTS idx_channels_name_unique 
    ON channels(name) WHERE deleted_at IS NULL;
