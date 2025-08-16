-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(255) PRIMARY KEY,
    channel_ids JSONB NOT NULL,
    variables JSONB NOT NULL,
    channel_overrides JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at BIGINT NOT NULL
);

-- Create message_results table
CREATE TABLE IF NOT EXISTS message_results (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(255) NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    channel_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    error_code VARCHAR(100),
    error_details TEXT,
    sent_at BIGINT
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_message_results_message_id ON message_results(message_id);
CREATE INDEX IF NOT EXISTS idx_message_results_channel_id ON message_results(channel_id);
CREATE INDEX IF NOT EXISTS idx_message_results_status ON message_results(status);

-- Add constraint for message status
ALTER TABLE messages ADD CONSTRAINT check_message_status 
    CHECK (status IN ('pending', 'success', 'failed', 'partial_success'));

-- Add constraint for message result status
ALTER TABLE message_results ADD CONSTRAINT check_message_result_status 
    CHECK (status IN ('success', 'failed'));

-- Add unique constraint to prevent duplicate results for same message and channel
CREATE UNIQUE INDEX IF NOT EXISTS idx_message_results_unique 
    ON message_results(message_id, channel_id);