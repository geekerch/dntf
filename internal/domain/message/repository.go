package message

import (
	"context"
)

// MessageRepository 訊息倉儲介面
type MessageRepository interface {
	// Save 儲存訊息
	Save(ctx context.Context, message *Message) error
	
	// FindByID 根據 ID 查找訊息
	FindByID(ctx context.Context, id *MessageID) (*Message, error)
	
	// Update 更新訊息
	Update(ctx context.Context, message *Message) error
	
	// Exists 檢查訊息是否存在
	Exists(ctx context.Context, id *MessageID) (bool, error)
}