package cqrs

import (
	"context"
	"time"
)

// Event represents a domain event in the CQRS pattern
type Event interface {
	// GetEventID returns the unique identifier for this event
	GetEventID() string
	// GetEventType returns the type of the event
	GetEventType() string
	// GetAggregateID returns the ID of the aggregate that generated this event
	GetAggregateID() string
	// GetAggregateType returns the type of the aggregate
	GetAggregateType() string
	// GetTimestamp returns when the event occurred
	GetTimestamp() time.Time
	// GetVersion returns the version of the aggregate when this event was generated
	GetVersion() int64
	// GetData returns the event data
	GetData() interface{}
}

// EventHandler handles a specific type of event
type EventHandler interface {
	// Handle processes the event
	Handle(ctx context.Context, event Event) error
	// GetEventType returns the type of event this handler processes
	GetEventType() string
}

// EventBus publishes and subscribes to events
type EventBus interface {
	// Publish publishes an event
	Publish(ctx context.Context, event Event) error
	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []Event) error
	// Subscribe subscribes to events of a specific type
	Subscribe(eventType string, handler EventHandler) error
	// Unsubscribe unsubscribes from events of a specific type
	Unsubscribe(eventType string, handler EventHandler) error
}

// EventStore stores and retrieves events
type EventStore interface {
	// SaveEvents saves events to the store
	SaveEvents(ctx context.Context, aggregateID string, events []Event, expectedVersion int64) error
	// GetEvents retrieves events for an aggregate
	GetEvents(ctx context.Context, aggregateID string, fromVersion int64) ([]Event, error)
	// GetAllEvents retrieves all events of a specific type
	GetAllEvents(ctx context.Context, eventType string, fromTimestamp time.Time) ([]Event, error)
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregateId"`
	AggregateType string                 `json:"aggregateType"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       int64                  `json:"version"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	UserID        string                 `json:"userId,omitempty"`
	TraceID       string                 `json:"traceId,omitempty"`
}

// GetEventID returns the event ID
func (e *BaseEvent) GetEventID() string {
	return e.ID
}

// GetEventType returns the event type
func (e *BaseEvent) GetEventType() string {
	return e.Type
}

// GetAggregateID returns the aggregate ID
func (e *BaseEvent) GetAggregateID() string {
	return e.AggregateID
}

// GetAggregateType returns the aggregate type
func (e *BaseEvent) GetAggregateType() string {
	return e.AggregateType
}

// GetTimestamp returns the event timestamp
func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion returns the aggregate version
func (e *BaseEvent) GetVersion() int64 {
	return e.Version
}

// GetData returns the event data
func (e *BaseEvent) GetData() interface{} {
	return e.Data
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, aggregateID, aggregateType string, version int64, data interface{}) *BaseEvent {
	return &BaseEvent{
		ID:            generateID(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Timestamp:     time.Now(),
		Version:       version,
		Data:          data,
		Metadata:      make(map[string]interface{}),
	}
}

// EventProjection represents a read model projection
type EventProjection interface {
	// GetProjectionName returns the name of the projection
	GetProjectionName() string
	// Handle processes an event to update the projection
	Handle(ctx context.Context, event Event) error
	// Reset resets the projection to its initial state
	Reset(ctx context.Context) error
	// GetLastProcessedVersion returns the last processed event version
	GetLastProcessedVersion(ctx context.Context) (int64, error)
	// SetLastProcessedVersion sets the last processed event version
	SetLastProcessedVersion(ctx context.Context, version int64) error
}

// ProjectionManager manages event projections
type ProjectionManager interface {
	// RegisterProjection registers a projection
	RegisterProjection(projection EventProjection) error
	// StartProjection starts processing events for a projection
	StartProjection(ctx context.Context, projectionName string) error
	// StopProjection stops processing events for a projection
	StopProjection(projectionName string) error
	// RebuildProjection rebuilds a projection from the beginning
	RebuildProjection(ctx context.Context, projectionName string) error
}