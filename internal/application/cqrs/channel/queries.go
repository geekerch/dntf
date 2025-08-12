package channel

import (
	"fmt"

	"notification/internal/application/cqrs"
)

// Query types
const (
	GetChannelQueryType  = "channel.get"
	ListChannelsQueryType = "channel.list"
)

// GetChannelQuery represents a query to get a single channel
type GetChannelQuery struct {
	*cqrs.BaseQuery
	ChannelID string `json:"channelId"`
}

// NewGetChannelQuery creates a new get channel query
func NewGetChannelQuery(channelID string) *GetChannelQuery {
	return &GetChannelQuery{
		BaseQuery: cqrs.NewBaseQuery(GetChannelQueryType),
		ChannelID: channelID,
	}
}

// Validate validates the get channel query
func (q *GetChannelQuery) Validate() error {
	if q.ChannelID == "" {
		return fmt.Errorf("channel ID is required")
	}
	
	return nil
}

// ListChannelsQuery represents a query to list channels
type ListChannelsQuery struct {
	*cqrs.BaseQuery
	ChannelType string             `json:"channelType,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Enabled     *bool              `json:"enabled,omitempty"`
	Options     *cqrs.QueryOptions `json:"options,omitempty"`
}

// NewListChannelsQuery creates a new list channels query
func NewListChannelsQuery() *ListChannelsQuery {
	return &ListChannelsQuery{
		BaseQuery: cqrs.NewBaseQuery(ListChannelsQueryType),
		Options:   cqrs.NewQueryOptions(),
	}
}

// WithChannelType sets the channel type filter
func (q *ListChannelsQuery) WithChannelType(channelType string) *ListChannelsQuery {
	q.ChannelType = channelType
	return q
}

// WithTags sets the tags filter
func (q *ListChannelsQuery) WithTags(tags []string) *ListChannelsQuery {
	q.Tags = tags
	return q
}

// WithEnabled sets the enabled filter
func (q *ListChannelsQuery) WithEnabled(enabled bool) *ListChannelsQuery {
	q.Enabled = &enabled
	return q
}

// WithPagination sets pagination options
func (q *ListChannelsQuery) WithPagination(offset, limit int) *ListChannelsQuery {
	q.Options.Pagination = &cqrs.Pagination{
		Offset: offset,
		Limit:  limit,
	}
	return q
}

// WithSorting adds sorting options
func (q *ListChannelsQuery) WithSorting(field, order string) *ListChannelsQuery {
	q.Options.Sorting = append(q.Options.Sorting, cqrs.Sorting{
		Field: field,
		Order: order,
	})
	return q
}

// WithFiltering adds filtering options
func (q *ListChannelsQuery) WithFiltering(field, operator string, value interface{}) *ListChannelsQuery {
	q.Options.Filtering = append(q.Options.Filtering, cqrs.Filtering{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// WithFields sets field selection
func (q *ListChannelsQuery) WithFields(fields []string) *ListChannelsQuery {
	q.Options.Fields = fields
	return q
}

// Validate validates the list channels query
func (q *ListChannelsQuery) Validate() error {
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