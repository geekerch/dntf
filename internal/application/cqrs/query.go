package cqrs

import (
	"context"
	"time"
)

// Query represents a query in the CQRS pattern
type Query interface {
	// GetQueryID returns the unique identifier for this query
	GetQueryID() string
	// GetQueryType returns the type of the query
	GetQueryType() string
	// GetTimestamp returns when the query was created
	GetTimestamp() time.Time
	// Validate validates the query
	Validate() error
}

// QueryResult represents the result of executing a query
type QueryResult struct {
	QueryID    string                 `json:"queryId"`
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data,omitempty"`
	Error      error                  `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ExecutedAt time.Time              `json:"executedAt"`
	Duration   time.Duration          `json:"duration"`
	CacheHit   bool                   `json:"cacheHit,omitempty"`
}

// QueryHandler handles a specific type of query
type QueryHandler interface {
	// Handle processes the query and returns the result
	Handle(ctx context.Context, query Query) (*QueryResult, error)
	// GetQueryType returns the type of query this handler processes
	GetQueryType() string
}

// QueryBus dispatches queries to their appropriate handlers
type QueryBus interface {
	// Execute executes a query
	Execute(ctx context.Context, query Query) (*QueryResult, error)
	// RegisterHandler registers a query handler
	RegisterHandler(handler QueryHandler) error
	// GetHandler returns the handler for a query type
	GetHandler(queryType string) (QueryHandler, error)
}

// BaseQuery provides common query functionality
type BaseQuery struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId,omitempty"`
	TraceID   string    `json:"traceId,omitempty"`
}

// GetQueryID returns the query ID
func (q *BaseQuery) GetQueryID() string {
	return q.ID
}

// GetQueryType returns the query type
func (q *BaseQuery) GetQueryType() string {
	return q.Type
}

// GetTimestamp returns the query timestamp
func (q *BaseQuery) GetTimestamp() time.Time {
	return q.Timestamp
}

// NewBaseQuery creates a new base query
func NewBaseQuery(queryType string) *BaseQuery {
	return &BaseQuery{
		ID:        generateID(),
		Type:      queryType,
		Timestamp: time.Now(),
	}
}

// Pagination represents pagination parameters for queries
type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Sorting represents sorting parameters for queries
type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// Filtering represents filtering parameters for queries
type Filtering struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "ne", "gt", "lt", "gte", "lte", "like", "in"
	Value    interface{} `json:"value"`
}

// QueryOptions represents common query options
type QueryOptions struct {
	Pagination *Pagination  `json:"pagination,omitempty"`
	Sorting    []Sorting    `json:"sorting,omitempty"`
	Filtering  []Filtering  `json:"filtering,omitempty"`
	Fields     []string     `json:"fields,omitempty"` // Field selection
	Include    []string     `json:"include,omitempty"` // Related entities to include
}

// NewQueryOptions creates new query options with defaults
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		Pagination: &Pagination{
			Offset: 0,
			Limit:  20,
		},
		Sorting:   []Sorting{},
		Filtering: []Filtering{},
		Fields:    []string{},
		Include:   []string{},
	}
}