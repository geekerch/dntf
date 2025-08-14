package message

import (
	"fmt"

	"notification/internal/application/cqrs"
)

// Query types
const (
	GetMessageQueryType  = "message.get"
	ListMessagesQueryType = "message.list"
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

// ListMessagesQuery represents a query to list messages
type ListMessagesQuery struct {
	*cqrs.BaseQuery
	ChannelID string             `json:"channelId,omitempty"`
	Status    string             `json:"status,omitempty"`
	Options   *cqrs.QueryOptions `json:"options,omitempty"`
}

// NewListMessagesQuery creates a new list messages query
func NewListMessagesQuery() *ListMessagesQuery {
	return &ListMessagesQuery{
		BaseQuery: cqrs.NewBaseQuery(ListMessagesQueryType),
		Options:   cqrs.NewQueryOptions(),
	}
}

// WithChannelID sets the channel ID filter
func (q *ListMessagesQuery) WithChannelID(channelID string) *ListMessagesQuery {
	q.ChannelID = channelID
	return q
}

// WithStatus sets the status filter
func (q *ListMessagesQuery) WithStatus(status string) *ListMessagesQuery {
	q.Status = status
	return q
}

// WithPagination sets pagination options
func (q *ListMessagesQuery) WithPagination(offset, limit int) *ListMessagesQuery {
	q.Options.Pagination = &cqrs.Pagination{
		Offset: offset,
		Limit:  limit,
	}
	return q
}

// WithSorting adds sorting options
func (q *ListMessagesQuery) WithSorting(field, order string) *ListMessagesQuery {
	q.Options.Sorting = append(q.Options.Sorting, cqrs.Sorting{
		Field: field,
		Order: order,
	})
	return q
}

// WithFiltering adds filtering options
func (q *ListMessagesQuery) WithFiltering(field, operator string, value interface{}) *ListMessagesQuery {
	q.Options.Filtering = append(q.Options.Filtering, cqrs.Filtering{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// WithFields sets field selection
func (q *ListMessagesQuery) WithFields(fields []string) *ListMessagesQuery {
	q.Options.Fields = fields
	return q
}

// Validate validates the list messages query
func (q *ListMessagesQuery) Validate() error {
	if q.Options != nil && q.Options.Pagination != nil {
		if q.Options.Pagination.Limit < 0 {
			return fmt.Errorf("pagination limit cannot be negative")
		}
		if q.Options.Pagination.Offset < 0 {
			return fmt.Errorf("pagination offset cannot be negative")
		}
		if q.Options.Pagination.Limit > 1000 {
			return fmt.Errorf("pagination limit cannot exceed 1000")
		}
	}
	
	// Validate sorting fields
	if q.Options != nil {
		for _, sort := range q.Options.Sorting {
			if sort.Field == "" {
				return fmt.Errorf("sorting field cannot be empty")
			}
			if sort.Order != "asc" && sort.Order != "desc" {
				return fmt.Errorf("sorting order must be 'asc' or 'desc'")
			}
		}
		
		// Validate filtering
		for _, filter := range q.Options.Filtering {
			if filter.Field == "" {
				return fmt.Errorf("filtering field cannot be empty")
			}
			validOperators := []string{"eq", "ne", "gt", "lt", "gte", "lte", "like", "in"}
			valid := false
			for _, op := range validOperators {
				if filter.Operator == op {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid filtering operator: %s", filter.Operator)
			}
		}
	}
	
	return nil
}