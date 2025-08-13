package template

import (
	"fmt"

	"notification/internal/application/cqrs"
)

// Query types
const (
	GetTemplateQueryType  = "template.get"
	ListTemplatesQueryType = "template.list"
)

// GetTemplateQuery represents a query to get a single template
type GetTemplateQuery struct {
	*cqrs.BaseQuery
	TemplateID string `json:"templateId"`
}

// NewGetTemplateQuery creates a new get template query
func NewGetTemplateQuery(templateID string) *GetTemplateQuery {
	return &GetTemplateQuery{
		BaseQuery:  cqrs.NewBaseQuery(GetTemplateQueryType),
		TemplateID: templateID,
	}
}

// Validate validates the get template query
func (q *GetTemplateQuery) Validate() error {
	if q.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}
	
	return nil
}

// ListTemplatesQuery represents a query to list templates
type ListTemplatesQuery struct {
	*cqrs.BaseQuery
	ChannelType string             `json:"channelType,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Options     *cqrs.QueryOptions `json:"options,omitempty"`
}

// NewListTemplatesQuery creates a new list templates query
func NewListTemplatesQuery() *ListTemplatesQuery {
	return &ListTemplatesQuery{
		BaseQuery: cqrs.NewBaseQuery(ListTemplatesQueryType),
		Options:   cqrs.NewQueryOptions(),
	}
}

// WithChannelType sets the channel type filter
func (q *ListTemplatesQuery) WithChannelType(channelType string) *ListTemplatesQuery {
	q.ChannelType = channelType
	return q
}

// WithTags sets the tags filter
func (q *ListTemplatesQuery) WithTags(tags []string) *ListTemplatesQuery {
	q.Tags = tags
	return q
}

// WithPagination sets pagination options
func (q *ListTemplatesQuery) WithPagination(offset, limit int) *ListTemplatesQuery {
	q.Options.Pagination = &cqrs.Pagination{
		Offset: offset,
		Limit:  limit,
	}
	return q
}

// WithSorting adds sorting options
func (q *ListTemplatesQuery) WithSorting(field, order string) *ListTemplatesQuery {
	q.Options.Sorting = append(q.Options.Sorting, cqrs.Sorting{
		Field: field,
		Order: order,
	})
	return q
}

// WithFiltering adds filtering options
func (q *ListTemplatesQuery) WithFiltering(field, operator string, value interface{}) *ListTemplatesQuery {
	q.Options.Filtering = append(q.Options.Filtering, cqrs.Filtering{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// WithFields sets field selection
func (q *ListTemplatesQuery) WithFields(fields []string) *ListTemplatesQuery {
	q.Options.Fields = fields
	return q
}

// Validate validates the list templates query
func (q *ListTemplatesQuery) Validate() error {
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