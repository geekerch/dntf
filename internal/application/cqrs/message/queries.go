package message

import (
	"fmt"

	"notification/internal/application/cqrs"
)

// Query types
const (
	GetMessageQueryType = "message.get"
)

// GetMessageQuery represents a query to get a single message
type GetMessageQuery struct {
	*cqrs.BaseQuery
	MessageID string `json:"messageId"`
}

// NewGetMessageQuery creates a new get message query
func NewGetMessageQuery(messageID string) *GetMessageQuery {
	return &GetMessageQuery{
		BaseQuery: cqrs.NewBaseQuery(GetMessageQueryType),
		MessageID: messageID,
	}
}

// Validate validates the get message query
func (q *GetMessageQuery) Validate() error {
	if q.MessageID == "" {
		return fmt.Errorf("message ID is required")
	}
	
	return nil
}